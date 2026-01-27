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

	"github.com/google/uuid"
	"github.com/rest-com-service/internal/client/resttcp"
	"github.com/rest-com-service/internal/repository"
	api "github.com/rest-com-service/pkg/models"
)

type RESTService struct {
	terminalCredRepo repository.TerminalCredentialRepository
	linkStateRepo    repository.H2HLinkStateRepository
	tcp              *resttcp.Client
	rng              *rand.Rand
}

func NewRESTService(
	terminalCredRepo repository.TerminalCredentialRepository,
	linkStateRepo repository.H2HLinkStateRepository,
	tcp *resttcp.Client,
) *RESTService {
	return &RESTService{
		terminalCredRepo: terminalCredRepo,
		linkStateRepo:    linkStateRepo,
		tcp:              tcp,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *RESTService) AuthorizeSale(ctx context.Context, req *api.AuthorizeSaleRequest) (*api.AuthorizeSaleResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("missing request")
	}
	hostType := "HOST2_REST"

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

	rrn := generateRRN12(s.rng)

	request := &api.AuthorizeSaleRequest{
		MessageType:  "sale",
		UserID:       userUUID.String(),
		CardNumber:   req.CardNumber,
		ExpiryMonth:  req.ExpiryMonth,
		ExpiryYear:   time.Now().Year(),
		CurrencyCode: req.CurrencyCode,
		Items:        req.Items,
		Amount:       req.Amount,
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	pretty, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return nil, err
	}

	log.Printf("AuthorizeSaleRequest:\n%s", pretty)

	response, err := s.tcp.Send(ctx, payload)
	if err != nil {
		return nil, err
	}

	var hostResp *api.HostResponse
	if err := json.Unmarshal(response, &hostResp); err != nil {
		log.Printf("Host response is not valid JSON: %v", err)
	} else {
		pretty, _ := json.MarshalIndent(hostResp, "", "  ")
		log.Printf("Host response:\n%s", pretty)
	}

	respBytes, err := json.Marshal(hostResp)
	if err != nil {
		return nil, err
	}

	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return &api.AuthorizeSaleResponse{
		HostType:       hostType,
		TID:            creds.TID,
		MID:            creds.MID,
		STAN:           stan,
		RRN:            rrn,
		Status:         hostResp.Status,
		ResponseCode:   hostResp.ResponseCode,
		AuthCode:       hostResp.AuthCode,
		RawRequestHex:  hex.EncodeToString(reqBytes),
		RawResponseHex: hex.EncodeToString(respBytes),
	}, nil
}

func (s *RESTService) AuthorizeVoid(ctx context.Context, req *api.AuthorizeVoidRequest) (*api.AuthorizeVoidResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("missing request")
	}

	userUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid userId")
	}

	hostType := "HOST2_REST"

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

	request := &api.AuthorizeVoidRequest{
		UserID:        userUUID.String(),
		TransactionID: req.TransactionID,
		MessageType:   "void",
	}

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	pretty, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return nil, err
	}
	log.Printf("AuthorizeVoidRequest:\n%s", pretty)

	response, err := s.tcp.Send(ctx, payload)
	if err != nil {
		return nil, err
	}

	var hostResp *api.HostResponse
	if err := json.Unmarshal(response, &hostResp); err != nil {
		log.Printf("Host response is not valid JSON: %v", err)
	} else {
		pretty, _ := json.MarshalIndent(hostResp, "", "  ")
		log.Printf("Host response:\n%s", pretty)
	}

	respBytes, err := json.Marshal(hostResp)
	if err != nil {
		return nil, err
	}

	reqBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return &api.AuthorizeVoidResponse{
		HostType:       hostType,
		TID:            creds.TID,
		MID:            creds.MID,
		STAN:           stan,
		Status:         hostResp.Status,
		ResponseCode:   hostResp.ResponseCode,
		RawRequestHex:  hex.EncodeToString(reqBytes),
		RawResponseHex: hex.EncodeToString(respBytes),
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
