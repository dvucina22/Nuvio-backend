package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/rest-host-service/internal/config"
	"github.com/rest-host-service/internal/server"
)

type SaleItem struct {
	Code      string `json:"code"`
	UnitPrice int64  `json:"unitPrice"`
	Quantity  int    `json:"quantity"`
}

type AuthorizeSaleRequest struct {
	MessageType  string     `json:"messageType"`
	UserID       string     `json:"userId"`
	CardNumber   string     `json:"cardNumber"`
	ExpiryMonth  int        `json:"expiryMonth"`
	ExpiryYear   int        `json:"expiryYear"`
	CurrencyCode string     `json:"currencyCode"`
	Amount       int64      `json:"amount"`
	Items        []SaleItem `json:"items"`
}

type AuthorizeVoidRequest struct {
	MessageType   string `json:"messageType"`
	UserID        string `json:"userId"`
	TransactionID int64  `json:"transactionId"`
}

type HostResponse struct {
	Status       string `json:"status"`
	ResponseCode string `json:"responseCode"`
	AuthCode     string `json:"authCode,omitempty"`
}

func main() {
	cfg := config.Load()

	server.Start(cfg.Addr, func(req []byte) ([]byte, error) {
		var envelope struct {
			MessageType string `json:"messageType"`
		}
		if err := json.Unmarshal(req, &envelope); err != nil {
			log.Printf("Invalid request JSON: %v", err)
			return nil, err
		}

		switch envelope.MessageType {
		case "sale":
			var saleReq AuthorizeSaleRequest
			if err := json.Unmarshal(req, &saleReq); err != nil {
				log.Printf("Invalid sale request JSON: %v", err)
				return nil, err
			}

			pretty, _ := json.MarshalIndent(saleReq, "", "  ")
			log.Printf("AuthorizeSaleRequest:\n%s", pretty)

			resp := HostResponse{
				Status:       "APPROVED",
				ResponseCode: "00",
				AuthCode:     "AA3310",
			}
			log.Printf("Response:\n%s", resp)
			return json.Marshal(resp)

		case "void":
			var voidReq AuthorizeVoidRequest
			if err := json.Unmarshal(req, &voidReq); err != nil {
				log.Printf("Invalid void request JSON: %v", err)
				return nil, err
			}

			pretty, _ := json.MarshalIndent(voidReq, "", "  ")
			log.Printf("AuthorizeVoidRequest:\n%s", pretty)

			resp := HostResponse{
				Status:       "APPROVED",
				ResponseCode: "00",
			}
			log.Printf("Response:\n%s", resp)
			return json.Marshal(resp)

		default:
			log.Printf("Unsupported messageType: %s", envelope.MessageType)
			return nil, fmt.Errorf("unsupported messageType: %s", envelope.MessageType)
		}
	})
}
