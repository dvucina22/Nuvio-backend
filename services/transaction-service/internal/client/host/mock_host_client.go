package host

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/transaction-service/internal/repository"
	"github.com/transaction-service/pkg/models"
)

type MockHostClient struct {
	terminalCredRepo repository.TerminalCredentialRepository
}

type HostClient interface {
	AuthorizeSale(ctx context.Context, userID uuid.UUID, req *models.HostSaleRequest) (*models.HostSaleResponse, error)
	AuthorizeVoid(ctx context.Context, userID uuid.UUID, orig *models.Transaction) (*models.HostSaleResponse, error)
}

func NewMockHostClient(terminalCredRepo repository.TerminalCredentialRepository) MockHostClient {
	return MockHostClient{
		terminalCredRepo: terminalCredRepo,
	}
}

func (c *MockHostClient) AuthorizeSale(ctx context.Context, userID uuid.UUID, req *models.HostSaleRequest) (*models.HostSaleResponse, error) {
	firstDigit, err := extractFirstDigit(req.CardNumber)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", models.ErrInvalidCardNumber, err)
	}
	hostType := resolveHostType(firstDigit)

	creds, err := c.terminalCredRepo.GetActiveByUserAndHost(ctx, userID, hostType)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to get terminal credentials: %v", models.ErrDatabaseOperation, err)
	}
	if creds == nil {
		return nil, fmt.Errorf("%w: user=%s host=%s",
			models.ErrTerminalCredentialsNotFound, userID, hostType)
	}

	now := time.Now().UTC()
	stan := generateSTAN()
	txTime := now.Format("150405")
	txDate := now.Format("0102")
	rrn := generateRRN(now)

	resp := &models.HostSaleResponse{
		HostType:        hostType,
		TerminalTID:     creds.TID,
		MerchantMID:     creds.MID,
		STAN:            stan,
		TransactionTime: txTime,
		TransactionDate: txDate,
		RRN:             rrn,
		Status:          "APPROVED",
		CardFirstDigit:  fmt.Sprintf("%d", firstDigit),
		RawRequest:      []byte("{}"),
		RawResponse:     []byte(`{"rc":"00"}`),
	}

	return resp, nil
}

func (c *MockHostClient) AuthorizeVoid(ctx context.Context, userID uuid.UUID, orig *models.Transaction) (*models.HostSaleResponse, error) {
	if orig == nil {
		return nil, fmt.Errorf("original transaction is nil")
	}

	hostType := orig.HostType

	now := time.Now().UTC()
	_ = now
	stan := generateSTAN()

	resp := &models.HostSaleResponse{
		HostType:        hostType,
		TerminalTID:     orig.TerminalTID,
		MerchantMID:     orig.MerchantMID,
		STAN:            stan,
		TransactionTime: orig.TransactionTime,
		TransactionDate: orig.TransactionDate,
		RRN:             orig.RRN,
		Status:          "APPROVED",
		CardFirstDigit:  orig.CardFirstDigit,
		RawRequest:      []byte("{}"),
		RawResponse:     []byte(`{"rc":"00"}`),
	}

	return resp, nil
}

func generateSTAN() string {
	n := rand.Intn(1_000_000)
	return fmt.Sprintf("%06d", n)
}

func generateRRN(t time.Time) string {
	prefix := t.Format("021504")
	n := rand.Intn(1_000_000)
	return prefix + fmt.Sprintf("%06d", n)
}

func extractFirstDigit(cardNumber string) (int, error) {
	for _, r := range cardNumber {
		if unicode.IsDigit(r) {
			return int(r - '0'), nil
		}
	}
	return 0, fmt.Errorf("%w: no digits in card number", models.ErrInvalidCardNumber)
}

func resolveHostType(firstDigit int) string {
	if firstDigit%2 == 0 {
		return "HOST2_REST"
	}
	return "HOST1_ISO8583"
}
