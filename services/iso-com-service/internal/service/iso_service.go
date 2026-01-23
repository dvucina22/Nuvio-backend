package service

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/iso-com-service/internal/client/isotcp"
	"github.com/iso-com-service/internal/iso8583"
	"github.com/iso-com-service/internal/repository"
	api "github.com/iso-com-service/pkg/models"
)

type ISOService struct {
	terminalCredRepo repository.TerminalCredentialRepository
	linkStateRepo    repository.H2HLinkStateRepository
	tcp              *isotcp.Client
	rng              *rand.Rand
}

func NewISOService(
	terminalCredRepo repository.TerminalCredentialRepository,
	linkStateRepo repository.H2HLinkStateRepository,
	tcp *isotcp.Client,
) *ISOService {
	return &ISOService{
		terminalCredRepo: terminalCredRepo,
		linkStateRepo:    linkStateRepo,
		tcp:              tcp,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *ISOService) AuthorizeSale(ctx context.Context, req *api.AuthorizeSaleRequest) (*api.AuthorizeSaleResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("missing request")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid userId")
	}

	if strings.TrimSpace(req.CardNumber) == "" {
		return nil, fmt.Errorf("cardNumber is required")
	}
	if req.ExpiryMonth < 1 || req.ExpiryMonth > 12 {
		return nil, fmt.Errorf("expiryMonth must be between 1 and 12")
	}
	if req.ExpiryYear < time.Now().Year() {
		return nil, fmt.Errorf("expiryYear is in the past")
	}
	if len(req.CurrencyCode) != 3 {
		return nil, fmt.Errorf("currencyCode must be 3 digits")
	}
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	firstDigit, err := extractFirstDigit(req.CardNumber)
	if err != nil {
		return nil, err
	}
	hostType := resolveHostType(firstDigit)

	if hostType != "HOST1_ISO8583" {
		return nil, fmt.Errorf("unsupported hostType for iso-com-service: %s", hostType)
	}

	creds, err := s.terminalCredRepo.GetActiveByUserAndHost(ctx, userUUID, hostType)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return nil, fmt.Errorf("terminal credentials not found")
	}

	stan, err := s.linkStateRepo.NextSTAN(ctx, hostType)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	rrn := generateRRN12(s.rng)

	productData, err := buildProductData(req.Items)
	if err != nil {
		return nil, err
	}

	msg1100, err := buildSale1100(req, creds.TID, creds.MID, stan, rrn, now, productData)
	if err != nil {
		return nil, err
	}

	wireReq := iso8583.Encode(msg1100)
	log.Printf("Iso com service sale request to host (hex): %X", wireReq)

	wireResp, err := s.tcp.Send(ctx, wireReq)
	if err != nil {
		return nil, err
	}

	respMsg, err := iso8583.Decode(wireResp)
	if err != nil {
		return nil, err
	}
	log.Printf("Host response (hex): %X", wireResp)

	jsonData, err := json.MarshalIndent(respMsg, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal request to JSON: %v", err)
	} else {
		log.Printf("Host response (decoded):\n%s", jsonData)
	}

	rc, auth, err := parseSale1110(respMsg)
	if err != nil {
		return nil, err
	}

	status := "DECLINED"
	if rc == "00" {
		status = "APPROVED"
	}

	return &api.AuthorizeSaleResponse{
		HostType:       hostType,
		TID:            creds.TID,
		MID:            creds.MID,
		STAN:           stan,
		RRN:            rrn,
		Status:         status,
		ResponseCode:   rc,
		AuthCode:       auth,
		RawRequestHex:  hex.EncodeToString(wireReq),
		RawResponseHex: hex.EncodeToString(wireResp),
	}, nil
}

func (s *ISOService) AuthorizeVoid(ctx context.Context, req *api.AuthorizeVoidRequest) (*api.AuthorizeVoidResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("missing request")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid userId")
	}

	origWire, err := hex.DecodeString(strings.TrimSpace(req.OriginalRequestHex))
	if err != nil {
		return nil, fmt.Errorf("invalid originalRequestHex")
	}

	origMsg, err := iso8583.Decode(origWire)
	if err != nil {
		return nil, fmt.Errorf("failed to decode original request")
	}
	if origMsg.MTI != "1100" {
		return nil, fmt.Errorf("original request is not 1100")
	}

	pan := origMsg.Fields[2]
	firstDigit, err := extractFirstDigit(string(pan))
	if err != nil {
		return nil, err
	}
	hostType := resolveHostType(firstDigit)

	if hostType != "HOST1_ISO8583" {
		return nil, fmt.Errorf("unsupported hostType for iso-com-service: %s", hostType)
	}

	creds, err := s.terminalCredRepo.GetActiveByUserAndHost(ctx, userUUID, hostType)
	if err != nil {
		return nil, err
	}
	if creds == nil {
		return nil, fmt.Errorf("terminal credentials not found")
	}

	stan, err := s.linkStateRepo.NextSTAN(ctx, hostType)
	if err != nil {
		return nil, err
	}

	msg1420, err := buildVoid1420(origMsg, creds.TID, creds.MID, stan)
	if err != nil {
		return nil, err
	}

	wireReq := iso8583.Encode(msg1420)
	log.Printf("Iso com service void request to host (hex): %X", wireReq)

	wireResp, err := s.tcp.Send(ctx, wireReq)
	if err != nil {
		return nil, err
	}
	log.Printf("Host response (hex): %X", wireResp)

	respMsg, err := iso8583.Decode(wireResp)
	if err != nil {
		return nil, err
	}

	rc, err := parseVoid1430(respMsg)
	if err != nil {
		return nil, err
	}

	status := "DECLINED"
	if rc == "00" {
		status = "APPROVED"
	}

	return &api.AuthorizeVoidResponse{
		HostType:       hostType,
		TID:            creds.TID,
		MID:            creds.MID,
		STAN:           stan,
		Status:         status,
		ResponseCode:   rc,
		RawRequestHex:  hex.EncodeToString(wireReq),
		RawResponseHex: hex.EncodeToString(wireResp),
	}, nil
}

var rrnAlphabet = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateRRN12(r *rand.Rand) string {
	b := make([]byte, 12)
	for i := 0; i < 12; i++ {
		b[i] = rrnAlphabet[r.Intn(len(rrnAlphabet))]
	}
	return string(b)
}

func extractFirstDigit(cardNumber string) (int, error) {
	for _, r := range cardNumber {
		if unicode.IsDigit(r) {
			return int(r - '0'), nil
		}
	}
	return 0, fmt.Errorf("no digits in card number")
}

func resolveHostType(firstDigit int) string {
	if firstDigit%2 == 0 {
		return "HOST2_REST"
	}
	return "HOST1_ISO8583"
}

func formatAmount12(cents int64) string {
	if cents < 0 {
		cents = 0
	}
	return fmt.Sprintf("%012d", cents)
}

func formatQty6(q int) string {
	if q < 0 {
		q = 0
	}
	return fmt.Sprintf("%06d", q)
}

func buildProductData(items []api.SaleItem) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("no items")
	}
	if len(items) > 10 {
		return "", fmt.Errorf("too many items (max 10)")
	}

	var b strings.Builder

	for _, it := range items {
		code := strings.TrimSpace(it.Code)
		if code == "" || len(code) > 32 {
			return "", fmt.Errorf("invalid item code")
		}
		if it.UnitPrice <= 0 || it.Quantity <= 0 {
			return "", fmt.Errorf("invalid item values")
		}

		b.WriteString(code)
		b.WriteByte('#')
		b.WriteString(formatAmount12(it.UnitPrice))
		b.WriteString(formatQty6(it.Quantity))
		b.WriteByte(';')
	}

	return b.String(), nil
}

func buildSale1100(req *api.AuthorizeSaleRequest, tid, mid, stan, rrn string, now time.Time, productData string) (*iso8583.Message, error) {
	yy := req.ExpiryYear % 100

	m := &iso8583.Message{
		MTI:    "1100",
		Fields: make(map[int][]byte),
	}
	cc, err := encodeN3BCD(req.CurrencyCode)
	if err != nil {
		return nil, err
	}

	m.SetField(2, []byte(req.CardNumber))
	m.SetField(3, []byte("000000"))
	m.SetField(4, []byte(formatAmount12(req.Amount)))
	m.SetField(11, []byte(stan))
	m.SetField(12, []byte(now.UTC().Format("150405")))
	m.SetField(13, []byte(now.UTC().Format("0102")))
	m.SetField(14, []byte(fmt.Sprintf("%02d%02d", yy, req.ExpiryMonth)))
	m.SetField(37, []byte(rrn))
	m.SetField(41, []byte(tid))
	m.SetField(42, []byte(mid))
	m.SetField(49, cc)
	m.SetField(63, []byte(productData))

	return m, nil
}

func buildVoid1420(orig *iso8583.Message, tid, mid, stan string) (*iso8583.Message, error) {
	m := &iso8583.Message{
		MTI:    "1420",
		Fields: make(map[int][]byte),
	}

	orig49, ok := orig.Fields[49]
	if !ok || len(orig49) != 2 {
		return nil, fmt.Errorf("missing/invalid DE49 in original request")
	}

	m.SetField(2, orig.Fields[2])
	m.SetField(3, []byte("020000"))
	m.SetField(4, orig.Fields[4])
	m.SetField(11, []byte(stan))
	m.SetField(12, orig.Fields[12])
	m.SetField(13, orig.Fields[13])
	m.SetField(14, orig.Fields[14])
	m.SetField(37, orig.Fields[37])
	m.SetField(41, []byte(tid))
	m.SetField(42, []byte(mid))
	m.SetField(49, orig49)

	return m, nil
}

func parseSale1110(resp *iso8583.Message) (string, string, error) {
	if resp == nil || resp.MTI != "1110" {
		return "", "", fmt.Errorf("invalid response MTI")
	}

	rcB, ok := resp.Fields[39]
	if !ok {
		return "", "", fmt.Errorf("missing DE39")
	}
	rc := string(rcB)

	auth := ""
	if rc == "00" {
		if a, ok := resp.Fields[38]; ok {
			auth = string(a)
		}
	}

	return rc, auth, nil
}

func parseVoid1430(resp *iso8583.Message) (string, error) {
	if resp == nil || resp.MTI != "1430" {
		return "", fmt.Errorf("invalid response MTI")
	}
	rcB, ok := resp.Fields[39]
	if !ok {
		return "", fmt.Errorf("missing DE39")
	}
	return string(rcB), nil
}
