package handler

import (
	"fmt"

	"github.com/iso-host-service/internal/iso8583"
)

func ValidateSaleRequest(msg *iso8583.Message) error {
	if msg == nil {
		return fmt.Errorf("nil message")
	}

	if msg.MTI != "1100" {
		return fmt.Errorf("invalid MTI: %s", msg.MTI)
	}

	mandatory := []int{2, 3, 4, 11, 12, 13, 14, 37, 41, 42, 49, 63}
	for _, f := range mandatory {
		v, ok := msg.Fields[f]
		if !ok || len(v) == 0 {
			return fmt.Errorf("missing field %d", f)
		}
	}

	if string(msg.Fields[3]) != "000000" {
		return fmt.Errorf("invalid processing code: %s", string(msg.Fields[3]))
	}

	pan := string(msg.Fields[2])
	if !hasAnyDigit(pan) {
		return fmt.Errorf("invalid pan")
	}

	if !isAllDigitsLen(msg.Fields[4], 12) {
		return fmt.Errorf("invalid amount")
	}

	if !isAllDigitsLen(msg.Fields[11], 6) {
		return fmt.Errorf("invalid stan")
	}

	if !isAllDigitsLen(msg.Fields[12], 6) {
		return fmt.Errorf("invalid time")
	}

	if !isAllDigitsLen(msg.Fields[13], 4) {
		return fmt.Errorf("invalid date")
	}

	if !isAllDigitsLen(msg.Fields[14], 4) {
		return fmt.Errorf("invalid expiry")
	}

	if len(msg.Fields[41]) != 8 {
		return fmt.Errorf("invalid tid")
	}

	if len(msg.Fields[42]) != 15 {
		return fmt.Errorf("invalid mid")
	}

	ccy, err := decodeCurrency49(msg.Fields[49])
	if err != nil {
		return fmt.Errorf("invalid currency")
	}

	if ccy != "978" {
		return fmt.Errorf("invalid currency")
	}

	if len(msg.Fields[63]) == 0 || len(msg.Fields[63]) > 520 {
		return fmt.Errorf("invalid field 63")
	}

	return nil
}
