package handler

import (
	"fmt"
	"unicode"
)

func isAllDigitsLen(b []byte, n int) bool {
	if len(b) != n {
		return false
	}
	for _, r := range string(b) {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func hasAnyDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func decodeCurrency49(b []byte) (string, error) {
	if len(b) != 2 {
		return "", fmt.Errorf("invalid length")
	}

	hiPad := int((b[0] >> 4) & 0x0F)
	d1 := int(b[0] & 0x0F)
	d2 := int((b[1] >> 4) & 0x0F)
	d3 := int(b[1] & 0x0F)

	if hiPad != 0 {
		return "", fmt.Errorf("invalid pad")
	}
	if d1 > 9 || d2 > 9 || d3 > 9 {
		return "", fmt.Errorf("invalid digits")
	}

	return fmt.Sprintf("%d%d%d", d1, d2, d3), nil
}
