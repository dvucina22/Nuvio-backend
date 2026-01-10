package service

import (
	"fmt"
)

func encodeN3BCD(s string) ([]byte, error) {
	if len(s) != 3 {
		return nil, fmt.Errorf("n3 must be 3 digits")
	}

	d0 := s[0]
	d1 := s[1]
	d2 := s[2]

	if d0 < '0' || d0 > '9' || d1 < '0' || d1 > '9' || d2 < '0' || d2 > '9' {
		return nil, fmt.Errorf("n3 must contain only digits")
	}

	b0 := byte(d0-'0') & 0x0F
	b1 := ((byte(d1-'0') & 0x0F) << 4) | (byte(d2-'0') & 0x0F)

	return []byte{b0, b1}, nil
}
