package service

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/transaction-service/internal/client/host"
	"github.com/transaction-service/internal/repository"
	"github.com/transaction-service/pkg/models"
)

type TransactionService struct {
	repo       repository.TransactionRepository
	hostClient host.MockHostClient
}

func NewTransactionService(repo repository.TransactionRepository, hostClient host.MockHostClient) *TransactionService {
	return &TransactionService{
		repo:       repo,
		hostClient: hostClient,
	}
}

func (s *TransactionService) CreateSale(ctx context.Context, userID string, req *models.SaleRequest) (*models.SaleResponse, error) {
	if req == nil {
		return nil, models.ErrMissingFields
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.ErrInvalidUserId
	}

	if err := validateSaleRequest(req); err != nil {
		return nil, err
	}

	txProducts, totalAmount, err := buildTransactionProductsFromRequest(req)
	if err != nil {
		return nil, err
	}

	if req.TotalAmount > 0 && req.TotalAmount != totalAmount {
		return nil, fmt.Errorf("%w: totalAmount (%d) does not match sum of products (%d)",
			models.ErrInvalidAmount, req.TotalAmount, totalAmount)
	}

	hostReq := &models.HostSaleRequest{
		CardNumber:   req.CardNumber,
		ExpiryMonth:  req.ExpiryMonth,
		ExpiryYear:   req.ExpiryYear,
		CurrencyCode: req.CurrencyCode,
		Amount:       totalAmount,
	}

	hostResp, err := s.hostClient.AuthorizeSale(ctx, userUUID, hostReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrDatabaseOperation, err)
	}

	panMasked := maskPAN(req.CardNumber)
	expYY := req.ExpiryYear % 100

	trx := &models.Transaction{
		UserID:                userUUID,
		BankCardID:            nil,
		Type:                  "SALE",
		Status:                hostResp.Status,
		PANMasked:             panMasked,
		CardFirstDigit:        hostResp.CardFirstDigit,
		CardExpirationYY:      fmt.Sprintf("%02d", expYY),
		CardExpirationMM:      fmt.Sprintf("%02d", req.ExpiryMonth),
		ProcessingCode:        "000000",
		Amount:                totalAmount,
		CurrencyCode:          req.CurrencyCode,
		STAN:                  hostResp.STAN,
		TransactionTime:       hostResp.TransactionTime,
		TransactionDate:       hostResp.TransactionDate,
		RRN:                   hostResp.RRN,
		TerminalTID:           hostResp.TerminalTID,
		MerchantMID:           hostResp.MerchantMID,
		HostType:              hostResp.HostType,
		OriginalTransactionID: nil,
		RequestPayload:        hostResp.RawRequest,
	}

	if err := s.repo.CreateSale(ctx, trx, txProducts); err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrDatabaseOperation, err)
	}

	saleProducts := make([]models.SaleProductResponse, 0, len(txProducts))

	for _, tp := range txProducts {
		sp := models.SaleProductResponse{
			ProductID: tp.ProductID,
			Quantity:  int32(tp.Quantity),
			UnitPrice: tp.UnitPrice,
			Name:      tp.ProductName,
			SKU:       tp.ProductSKU,
		}
		saleProducts = append(saleProducts, sp)
	}

	saleResp := &models.SaleResponse{
		ID:           trx.ID,
		Status:       trx.Status,
		Amount:       trx.Amount,
		CurrencyCode: trx.CurrencyCode,
		CreatedAt:    trx.CreatedAt,
		Products:     saleProducts,
	}

	return saleResp, nil
}

func (s *TransactionService) VoidSale(ctx context.Context, req *models.VoidRequest) (*models.Transaction, error) {
	if req == nil || req.TransactionID <= 0 {
		return nil, models.ErrMissingFields
	}

	orig, err := s.repo.GetTransactionByIDAdmin(ctx, req.TransactionID)
	if err != nil {
		return nil, models.ErrDatabaseOperation
	}
	if orig == nil {
		return nil, models.ErrTransactionNotFound
	}

	if strings.ToUpper(orig.Type) != "SALE" {
		return nil, models.ErrInvalidTransactionType
	}
	if strings.ToUpper(orig.Status) != "APPROVED" {
		return nil, models.ErrInvalidTransactionState
	}

	hasVoid, err := s.repo.HasVoidForOriginal(ctx, orig.ID)
	if err != nil {
		return nil, fmt.Errorf("VoidTransaction failed: %w", err)
	}
	if hasVoid {
		return nil, models.ErrVoidAlreadyExists
	}

	hostResp, err := s.hostClient.AuthorizeVoid(ctx, orig)
	if err != nil {
		return nil, fmt.Errorf("Void transaction failed: %w", err)
	}

	voidTx := &models.Transaction{
		UserID: orig.UserID,

		BankCardID:            orig.BankCardID,
		Type:                  "VOID",
		Status:                hostResp.Status,
		PANMasked:             orig.PANMasked,
		CardFirstDigit:        hostResp.CardFirstDigit,
		CardExpirationYY:      orig.CardExpirationYY,
		CardExpirationMM:      orig.CardExpirationMM,
		ProcessingCode:        "020000",
		Amount:                orig.Amount,
		CurrencyCode:          orig.CurrencyCode,
		STAN:                  hostResp.STAN,
		TransactionTime:       hostResp.TransactionTime,
		TransactionDate:       hostResp.TransactionDate,
		RRN:                   hostResp.RRN,
		TerminalTID:           hostResp.TerminalTID,
		MerchantMID:           hostResp.MerchantMID,
		HostType:              hostResp.HostType,
		OriginalTransactionID: &orig.ID,
		RequestPayload:        hostResp.RawRequest,
	}

	if err := s.repo.VoidTransaction(ctx, voidTx); err != nil {
		return nil, models.ErrDatabaseOperation
	}

	return voidTx, nil
}

func (s *TransactionService) GetTransactionDetail(ctx context.Context, userID string, txID int64) (*models.Transaction, []*models.TransactionProduct, error) {
	if txID <= 0 {
		return nil, nil, models.ErrMissingFields
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, models.ErrInvalidUserId
	}

	tx, products, err := s.repo.GetTransactionWithProducts(ctx, userUUID, txID)
	if err != nil {
		return nil, nil, models.ErrDatabaseOperation
	}
	if tx == nil {
		return nil, nil, models.ErrTransactionNotFound
	}

	return tx, products, nil
}

func buildTransactionProductsFromRequest(req *models.SaleRequest) ([]*models.TransactionProduct, int64, error) {
	var (
		result []*models.TransactionProduct
		total  int64
	)

	for _, p := range req.Products {
		if p.UnitPrice <= 0 {
			return nil, 0, fmt.Errorf("%w: invalid unitPrice for productId=%d",
				models.ErrInvalidProducts, p.ProductID)
		}

		lineAmount := p.UnitPrice * int64(p.Quantity)
		total += lineAmount

		tp := &models.TransactionProduct{
			ProductID: p.ProductID,
			UnitPrice: p.UnitPrice,
			Quantity:  p.Quantity,
		}

		if p.Name != nil {
			name := *p.Name
			tp.ProductName = &name
		}
		if p.SKU != nil {
			sku := *p.SKU
			tp.ProductSKU = &sku
		}

		result = append(result, tp)
	}

	return result, total, nil
}

func validateSaleRequest(req *models.SaleRequest) error {
	if strings.TrimSpace(req.CardNumber) == "" {
		return fmt.Errorf("%w: cardNumber is required", models.ErrMissingFields)
	}
	if req.ExpiryMonth < 1 || req.ExpiryMonth > 12 {
		return fmt.Errorf("%w: expiryMonth must be between 1 and 12", models.ErrInvalidCard)
	}
	if req.ExpiryYear < time.Now().Year() {
		return fmt.Errorf("%w: expiryYear is in the past", models.ErrCardExpired)
	}
	if len(req.CurrencyCode) != 3 {
		return fmt.Errorf("%w: currencyCode must be 3 chars (ISO 4217 numeric)", models.ErrInvalidCurrencyCode)
	}
	if len(req.Products) == 0 {
		return fmt.Errorf("%w: at least one product is required", models.ErrInvalidProducts)
	}
	for _, p := range req.Products {
		if p.ProductID <= 0 {
			return fmt.Errorf("%w: invalid productId=%d", models.ErrInvalidProducts, p.ProductID)
		}
		if p.Quantity <= 0 {
			return fmt.Errorf("%w: invalid quantity for productId=%d", models.ErrInvalidProducts, p.ProductID)
		}
		if p.UnitPrice <= 0 {
			return fmt.Errorf("%w: invalid unitPrice for productId=%d", models.ErrInvalidProducts, p.ProductID)
		}
	}
	return nil
}

func maskPAN(pan string) string {
	var digits []rune
	for _, r := range pan {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}
	n := len(digits)
	if n <= 10 {
		for i := 0; i < n-4; i++ {
			digits[i] = '*'
		}
		return string(digits)
	}

	for i := 6; i < n-4; i++ {
		digits[i] = '*'
	}
	return string(digits)
}
