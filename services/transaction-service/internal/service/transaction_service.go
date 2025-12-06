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

func NewTransactionService(
	repo repository.TransactionRepository,
	hostClient host.MockHostClient,
) *TransactionService {
	return &TransactionService{
		repo:       repo,
		hostClient: hostClient,
	}
}

func (s *TransactionService) CreateSale(
	ctx context.Context,
	userID string,
	req *models.SaleRequest,
) (*models.Transaction, []*models.TransactionProduct, error) {
	if req == nil {
		return nil, nil, models.ErrMissingFields
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, nil, models.ErrInvalidUserId
	}

	if err := validateSaleRequest(req); err != nil {
		return nil, nil, err
	}

	txProducts, totalAmount, err := buildTransactionProductsFromRequest(req)
	if err != nil {
		return nil, nil, err
	}

	if req.TotalAmount > 0 && req.TotalAmount != totalAmount {
		return nil, nil, fmt.Errorf("%w: totalAmount (%d) does not match sum of products (%d)",
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
		return nil, nil, fmt.Errorf("%w: %v", models.ErrDatabaseOperation, err)
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
		return nil, nil, fmt.Errorf("%w: %v", models.ErrDatabaseOperation, err)
	}

	return trx, txProducts, nil
}

func buildTransactionProductsFromRequest(
	req *models.SaleRequest,
) ([]*models.TransactionProduct, int64, error) {
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
