package iso8583

import "fmt"

func DecodeCurrencyN3(b []byte) (string, error) {
	if len(b) != 2 {
		return "", fmt.Errorf("currency n3 must be 2 bytes, got %d", len(b))
	}

	d1 := int(b[0] & 0x0F)
	d2 := int((b[1] >> 4) & 0x0F)
	d3 := int(b[1] & 0x0F)

	if d1 > 9 || d2 > 9 || d3 > 9 {
		return "", fmt.Errorf("invalid currency BCD: %02X%02X", b[0], b[1])
	}

	return fmt.Sprintf("%d%d%d", d1, d2, d3), nil
}

func EncodeCurrencyN3(code string) ([]byte, error) {
	if len(code) != 3 {
		return nil, fmt.Errorf("currency must be 3 digits, got %q", code)
	}

	d1 := code[0] - '0'
	d2 := code[1] - '0'
	d3 := code[2] - '0'

	if d1 > 9 || d2 > 9 || d3 > 9 {
		return nil, fmt.Errorf("currency must be digits, got %q", code)
	}

	return []byte{
		byte(0x00 | (d1 & 0x0F)),
		byte(((d2 & 0x0F) << 4) | (d3 & 0x0F)),
	}, nil
}
