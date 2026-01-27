package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	cardSvc    *CardService
	hostClient host.ComClient
}

func NewTransactionService(repo repository.TransactionRepository, hostClient host.ComClient, cardSvc *CardService) *TransactionService {
	return &TransactionService{
		repo:       repo,
		hostClient: hostClient,
		cardSvc:    cardSvc,
	}
}

func (s *TransactionService) CreateSale(ctx context.Context, userID string, req *models.SaleRequest) (*models.SaleCreateResponse, error) {
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

	var (
		cardPAN       string
		expMonth      int
		expYear       int
		bankCardIDPtr *int64
	)

	if req.CardID != nil {
		paymentCard, err := s.cardSvc.GetCardForPayment(ctx, userID, int(*req.CardID))
		if err != nil {
			return nil, err
		}

		cardPAN = paymentCard.PAN
		expMonth = paymentCard.ExpirationMonth
		expYear = paymentCard.ExpirationYear

		id64 := int64(*req.CardID)
		bankCardIDPtr = &id64
	} else {
		cardPAN = req.CardNumber
		expMonth = req.ExpiryMonth
		expYear = req.ExpiryYear
	}

	txProducts, totalAmount, err := buildTransactionProductsFromRequest(req)
	if err != nil {
		return nil, err
	}

	if req.TotalAmount > 0 && req.TotalAmount != totalAmount {
		return nil, fmt.Errorf(
			"%w: totalAmount (%d) does not match sum of products (%d)",
			models.ErrInvalidAmount,
			req.TotalAmount,
			totalAmount,
		)
	}

	hostReq := &models.HostSaleRequest{
		CardNumber:   cardPAN,
		ExpiryMonth:  expMonth,
		ExpiryYear:   expYear,
		CurrencyCode: req.CurrencyCode,
		Amount:       totalAmount,
		Products:     txProducts,
	}

	hostResp, err := s.hostClient.AuthorizeSale(ctx, userUUID, hostReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrDatabaseOperation, err)
	}

	jsonData, err := json.MarshalIndent(hostResp, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal hostResp to JSON: %v", err)
	} else {
		log.Printf("Com client response:\n%s", jsonData)
	}

	panMasked := maskPAN(cardPAN)
	expYY := expYear % 100

	trx := &models.Transaction{
		UserID:                userUUID,
		BankCardID:            bankCardIDPtr,
		Type:                  "SALE",
		Status:                hostResp.Status,
		PANMasked:             panMasked,
		CardFirstDigit:        hostResp.CardFirstDigit,
		CardExpirationYY:      fmt.Sprintf("%02d", expYY),
		CardExpirationMM:      fmt.Sprintf("%02d", expMonth),
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
		ResponsePayload:       hostResp.RawResponse,
		ResponseCode:          hostResp.ResponseCode,
		AuthCode:              hostResp.AuthCode,
	}

	if err := s.repo.CreateSale(ctx, trx, txProducts); err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrDatabaseOperation, err)
	}

	saleResp := &models.SaleCreateResponse{
		ID:           trx.ID,
		Status:       trx.Status,
		ResponseCode: *trx.ResponseCode,
		AuthCode:     trx.AuthCode,
		CreatedAt:    trx.CreatedAt,
	}

	return saleResp, nil
}

func (s *TransactionService) VoidSale(ctx context.Context, req *models.VoidRequest) (*models.VoidCreateResponse, error) {
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
		return nil, fmt.Errorf("%w: %v", models.ErrDatabaseOperation, err)
	}

	jsonData, err := json.MarshalIndent(hostResp, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal hostResp to JSON: %v", err)
	} else {
		log.Printf("Com client response:\n%s", jsonData)
	}

	var rcPtr *string
	if hostResp.ResponseCode != nil && strings.TrimSpace(*hostResp.ResponseCode) != "" {
		rc := *hostResp.ResponseCode
		rcPtr = &rc
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
		ResponsePayload:       hostResp.RawResponse,
		ResponseCode:          rcPtr,
		AuthCode:              hostResp.AuthCode,
	}

	if err := s.repo.VoidTransaction(ctx, voidTx); err != nil {
		return nil, models.ErrDatabaseOperation
	}

	voidResp := &models.VoidCreateResponse{
		ID:                    voidTx.ID,
		OriginalTransactionID: orig.ID,
		Status:                voidTx.Status,
		ResponseCode:          *voidTx.ResponseCode,
		CreatedAt:             voidTx.CreatedAt,
	}

	return voidResp, nil
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
			ProductID:  p.ProductID,
			UnitPrice:  p.UnitPrice,
			Quantity:   p.Quantity,
			ProductSKU: p.SKU,
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
	if req.CardID == nil {
		if strings.TrimSpace(req.CardNumber) == "" {
			return fmt.Errorf("%w: cardNumber is required", models.ErrMissingFields)
		}
		if req.ExpiryMonth < 1 || req.ExpiryMonth > 12 {
			return fmt.Errorf("%w: expiryMonth must be between 1 and 12", models.ErrInvalidCard)
		}
		if req.ExpiryYear < time.Now().Year() {
			return fmt.Errorf("%w: expiryYear is in the past", models.ErrCardExpired)
		}
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

func (s *TransactionService) GetFilteredTransactions(ctx context.Context, userID string, filter *models.TransactionFilter, isAdmin bool) (*models.TransactionListResponse, error) {
	if err := filter.Validate(); err != nil {
		return nil, err
	}

	if isAdmin {
		adminFilter := &models.AdminTransactionFilter{
			Search:          filter.Search,
			DateFrom:        filter.DateFrom,
			DateTo:          filter.DateTo,
			Statuses:        filter.Statuses,
			Types:           filter.Types,
			AmountMin:       filter.AmountMin,
			AmountMax:       filter.AmountMax,
			ProductCountMin: filter.ProductCountMin,
			ProductCountMax: filter.ProductCountMax,
			UserID:          filter.UserID,
			Page:            filter.Page,
			PageSize:        filter.PageSize,
		}

		adminTransactions, total, err := s.repo.GetFilteredAdminTransactions(ctx, adminFilter)
		if err != nil {
			return nil, models.ErrDatabaseOperation
		}

		userTransactions := make([]*models.TransactionListItem, len(adminTransactions))
		for i, tx := range adminTransactions {
			userTransactions[i] = &models.TransactionListItem{
				ID:           tx.ID,
				Status:       tx.Status,
				Amount:       tx.Amount,
				CurrencyCode: tx.CurrencyCode,
				PANMasked:    tx.PANMasked,
				ProductCount: tx.ProductCount,
				ProductIds:   tx.ProductIds,
				CreatedAt:    tx.CreatedAt,
			}
		}

		return &models.TransactionListResponse{
			Transactions: userTransactions,
			Total:        total,
			Page:         adminFilter.Page,
			PageSize:     adminFilter.PageSize,
		}, nil
	}

	filter.UserID = &userID

	transactions, total, err := s.repo.GetFilteredUserTransactions(ctx, filter)
	if err != nil {
		return nil, models.ErrDatabaseOperation
	}

	return &models.TransactionListResponse{
		Transactions: transactions,
		Total:        total,
		Page:         filter.Page,
		PageSize:     filter.PageSize,
	}, nil
}

func (s *TransactionService) GetTransactionDetails(ctx context.Context, userID string, txID int64, isAdmin bool) (interface{}, error) {
	if txID <= 0 {
		return nil, models.ErrMissingFields
	}

	if isAdmin {
		detail, err := s.repo.GetAdminTransactionDetail(ctx, txID)
		if err != nil {
			return nil, models.ErrDatabaseOperation
		}
		if detail == nil {
			return nil, models.ErrTransactionNotFound
		}
		return detail, nil
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return nil, models.ErrInvalidUserId
	}

	detail, err := s.repo.GetUserTransactionDetail(ctx, userUUID, txID)
	if err != nil {
		return nil, models.ErrDatabaseOperation
	}
	if detail == nil {
		return nil, models.ErrTransactionNotFound
	}

	return detail, nil
}

func (s *TransactionService) GetTransactionStatistics(ctx context.Context) (*models.TransactionStatistics, error) {
	stats, err := s.repo.GetTransactionStatistics(ctx)
	if err != nil {
		return nil, models.ErrDatabaseOperation
	}

	return stats, nil
}
