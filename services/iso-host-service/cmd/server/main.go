package main

import (
	"encoding/json"
	"log"

	"github.com/iso-host-service/internal/config"
	"github.com/iso-host-service/internal/handler"
	"github.com/iso-host-service/internal/iso8583"
	"github.com/iso-host-service/internal/server"
)

func main() {
	cfg := config.Load()

	server.Start(cfg.Addr, func(req []byte) ([]byte, error) {
		log.Printf("Host received request (hex): %X", req)
		msg, err := iso8583.Decode(req)

		jsonData, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			log.Printf("Failed to marshal request to JSON: %v", err)
		} else {
			log.Printf("Decoded ISO8583 message:\n%s", jsonData)
		}

		if err != nil {
			return iso8583.Encode(iso8583.FormatErrorResponse("1100")), nil
		}

		switch msg.MTI {
		case "1100":
			if err := handler.ValidateSaleRequest(msg); err != nil {
				resp := iso8583.NewResponse(msg)
				resp.SetField(39, []byte("30"))
				resp.RemoveField(38)
				return iso8583.Encode(resp), nil
			}

			resp, _ := handler.HandleSale(msg)
			return iso8583.Encode(resp), nil

		case "1420":
			if b, ok := msg.Fields[49]; ok {
				log.Printf("DE49 len=%d hex=%x", len(b), b)
			} else {
				log.Printf("DE49 missing")
			}

			if err := handler.ValidateVoidRequest(msg); err != nil {
				log.Printf("VOID validation failed: %v", err)
				resp := iso8583.NewResponse(msg)
				resp.SetField(39, []byte("30"))
				resp.RemoveField(38)
				return iso8583.Encode(resp), nil
			}

			resp, _ := handler.HandleVoid(msg)
			return iso8583.Encode(resp), nil

		default:
			resp := iso8583.NewResponse(msg)
			resp.SetField(39, []byte("30"))
			resp.RemoveField(38)
			return iso8583.Encode(resp), nil
		}
	})
}
