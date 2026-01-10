package handler

import (
	"fmt"

	"github.com/iso-host-service/internal/iso8583"
)

func ValidateVoidRequest(msg *iso8583.Message) error {
	if msg == nil {
		return fmt.Errorf("nil message")
	}
	if msg.MTI != "1420" {
		return fmt.Errorf("invalid mti")
	}

	required := []int{2, 3, 4, 11, 12, 13, 14, 37, 41, 42, 49}
	for _, f := range required {
		v, ok := msg.Fields[f]
		if !ok || len(v) == 0 {
			return fmt.Errorf("missing field %d", f)
		}
	}

	if string(msg.Fields[3]) != "020000" {
		return fmt.Errorf("invalid processing code")
	}

	pan := string(msg.Fields[2])
	if !hasAnyDigit(pan) {
		return fmt.Errorf("invalid pan")
	}

	ccy, err := decodeCurrency49(msg.Fields[49])
	if err != nil {
		return fmt.Errorf("invalid currency")
	}

	if ccy != "978" {
		return fmt.Errorf("invalid currency")
	}

	return nil
}
