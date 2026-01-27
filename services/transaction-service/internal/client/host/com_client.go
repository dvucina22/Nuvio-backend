package host

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/transaction-service/pkg/models"
)

type HostClient interface {
	AuthorizeSale(ctx context.Context, userID uuid.UUID, req *models.HostSaleRequest) (*models.HostSaleResponse, error)
	AuthorizeVoid(ctx context.Context, orig *models.Transaction) (*models.HostSaleResponse, error)
}

type TokenProvider func(ctx context.Context) (string, error)

type ComClient struct {
	isoBaseURL    string
	restBaseURL   string
	httpClient    *http.Client
	tokenProvider TokenProvider
}

func NewComClient(isoBaseURL string, restBaseURL string, tokenProvider TokenProvider) *ComClient {
	return &ComClient{
		isoBaseURL:  strings.TrimRight(isoBaseURL, "/"),
		restBaseURL: strings.TrimRight(restBaseURL, "/"),
		httpClient: &http.Client{
			Timeout: 12 * time.Second,
		},
		tokenProvider: tokenProvider,
	}
}

func (c *ComClient) AuthorizeSale(ctx context.Context, userID uuid.UUID, req *models.HostSaleRequest) (*models.HostSaleResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("missing request")
	}
	if strings.TrimSpace(req.CardNumber) == "" {
		return nil, fmt.Errorf("cardNumber is required")
	}
	if req.ExpiryMonth < 1 || req.ExpiryMonth > 12 {
		return nil, fmt.Errorf("expiryMonth out of range")
	}
	if req.ExpiryYear <= 0 {
		return nil, fmt.Errorf("expiryYear is required")
	}
	if len(req.CurrencyCode) != 3 {
		return nil, fmt.Errorf("currencyCode must be 3 digits")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}

	if len(req.Products) == 0 {
		return nil, fmt.Errorf("products are required")
	}

	items := make([]models.SaleItemDTO, 0, len(req.Products))
	for _, p := range req.Products {
		if p == nil {
			return nil, fmt.Errorf("nil product")
		}
		if p.UnitPrice <= 0 || p.Quantity <= 0 {
			return nil, fmt.Errorf("invalid product values")
		}

		code := fmt.Sprintf("%d", p.ProductID)
		if p.ProductSKU != nil && strings.TrimSpace(*p.ProductSKU) != "" {
			code = strings.TrimSpace(*p.ProductSKU)
		}

		items = append(items, models.SaleItemDTO{
			Code:      code,
			UnitPrice: p.UnitPrice,
			Quantity:  p.Quantity,
		})
	}

	in := models.AuthorizeSaleReqDTO{
		UserID:       userID.String(),
		CardNumber:   req.CardNumber,
		ExpiryMonth:  req.ExpiryMonth,
		ExpiryYear:   req.ExpiryYear,
		CurrencyCode: req.CurrencyCode,
		Amount:       req.Amount,
		Items:        items,
	}

	cardFirstDigit := -1
	for _, r := range req.CardNumber {
		if r >= '0' && r <= '9' {
			cardFirstDigit = int(r - '0')
			break
		}
	}

	if cardFirstDigit == -1 {
		return nil, fmt.Errorf("card number contains no digits")
	}

	baseURL := c.restBaseURL
	if cardFirstDigit%2 == 1 {
		baseURL = c.isoBaseURL
	}

	var out models.AuthorizeSaleRespDTO
	if err := c.post(ctx, baseURL+"/authorize/sale", in, &out); err != nil {
		return nil, err
	}

	cardFirstDigitStr := strconv.Itoa(cardFirstDigit)

	rawReq, err := decodeHexBytes(out.RawRequestHex)
	if err != nil {
		return nil, fmt.Errorf("invalid RawRequestHex: %w", err)
	}

	rawResp, err := decodeHexBytes(out.RawResponseHex)
	if err != nil {
		return nil, fmt.Errorf("invalid RawResponseHex: %w", err)
	}

	var authPtr *string
	if strings.TrimSpace(out.AuthCode) != "" {
		a := strings.TrimSpace(out.AuthCode)
		authPtr = &a
	}

	now := time.Now().UTC()

	var rcPtr *string
	if strings.TrimSpace(out.ResponseCode) != "" {
		rc := strings.TrimSpace(out.ResponseCode)
		rcPtr = &rc
	}

	return &models.HostSaleResponse{
		HostType:        out.HostType,
		TerminalTID:     out.TID,
		MerchantMID:     out.MID,
		STAN:            out.STAN,
		TransactionTime: now.Format("150405"),
		TransactionDate: now.Format("0102"),
		RRN:             out.RRN,
		Status:          out.Status,
		CardFirstDigit:  cardFirstDigitStr,
		RawRequest:      rawReq,
		RawResponse:     rawResp,
		ResponseCode:    rcPtr,
		AuthCode:        authPtr,
	}, nil
}

func (c *ComClient) AuthorizeVoid(ctx context.Context, orig *models.Transaction) (*models.HostSaleResponse, error) {
	if orig == nil {
		return nil, fmt.Errorf("original transaction is nil")
	}
	if len(orig.RequestPayload) == 0 {
		return nil, fmt.Errorf("original RequestPayload missing")
	}

	in := models.AuthorizeVoidReqDTO{
		UserID:             orig.UserID.String(),
		OriginalRequestHex: hex.EncodeToString(orig.RequestPayload),
	}

	firstDigit, err := strconv.Atoi(orig.CardFirstDigit)
	if err != nil {
		return nil, fmt.Errorf("invalid card first digit: %q", orig.CardFirstDigit)
	}

	baseURL := c.restBaseURL
	if firstDigit%2 == 1 {
		baseURL = c.isoBaseURL
	}

	var out models.AuthorizeVoidRespDTO
	if err := c.post(ctx, baseURL+"/api/bank-comm/authorize/void", in, &out); err != nil {
		return nil, err
	}

	rawReq, err := decodeHexBytes(out.RawRequestHex)
	if err != nil {
		return nil, fmt.Errorf("invalid RawRequestHex: %w", err)
	}

	rawResp, err := decodeHexBytes(out.RawResponseHex)
	if err != nil {
		return nil, fmt.Errorf("invalid RawResponseHex: %w", err)
	}

	var rcPtr *string
	if strings.TrimSpace(out.ResponseCode) != "" {
		rc := strings.TrimSpace(out.ResponseCode)
		rcPtr = &rc
	}

	return &models.HostSaleResponse{
		HostType:        out.HostType,
		TerminalTID:     out.TID,
		MerchantMID:     out.MID,
		STAN:            out.STAN,
		TransactionTime: orig.TransactionTime,
		TransactionDate: orig.TransactionDate,
		RRN:             orig.RRN,
		Status:          out.Status,
		CardFirstDigit:  orig.CardFirstDigit,
		RawRequest:      rawReq,
		RawResponse:     rawResp,
		ResponseCode:    rcPtr,
	}, nil
}

func (c *ComClient) post(ctx context.Context, path string, in any, out any) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	if authHeader, ok := AuthHeaderFromContext(ctx); ok {
		req.Header.Set("Authorization", authHeader)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	bodyBytes, _ := io.ReadAll(res.Body)

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("comm returned %d: %s", res.StatusCode, string(bodyBytes))
	}

	if out == nil {
		return nil
	}

	if err := json.Unmarshal(bodyBytes, out); err != nil {
		return fmt.Errorf("failed to decode comm response: %w", err)
	}

	return nil
}

func decodeHexBytes(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")
	s = strings.TrimPrefix(s, "\\x")
	s = strings.TrimPrefix(s, "\\X")
	s = strings.TrimSpace(s)

	if s == "" {
		return nil, fmt.Errorf("empty hex string")
	}

	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return b, nil
}
