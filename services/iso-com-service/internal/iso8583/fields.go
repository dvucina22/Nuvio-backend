package iso8583

import (
	"fmt"
)

type fieldKind int

const (
	kindFixedASCII fieldKind = iota
	kindLVARASCII
	kindLLVARASCII
	kindN3BCD
)

type fieldSpec struct {
	kind   fieldKind
	fixed  int
	maxLen int
}

var specs = map[int]fieldSpec{
	2:  {kind: kindLVARASCII, maxLen: 19},
	3:  {kind: kindFixedASCII, fixed: 6},
	4:  {kind: kindFixedASCII, fixed: 12},
	11: {kind: kindFixedASCII, fixed: 6},
	12: {kind: kindFixedASCII, fixed: 6},
	13: {kind: kindFixedASCII, fixed: 4},
	14: {kind: kindFixedASCII, fixed: 4},
	37: {kind: kindFixedASCII, fixed: 12},
	38: {kind: kindFixedASCII, fixed: 6},
	39: {kind: kindFixedASCII, fixed: 2},
	41: {kind: kindFixedASCII, fixed: 8},
	42: {kind: kindFixedASCII, fixed: 15},
	49: {kind: kindN3BCD, fixed: 2},
	63: {kind: kindLLVARASCII, maxLen: 520},
}

func decodeField(bit int, data []byte) ([]byte, int) {
	spec, ok := specs[bit]
	if !ok {
		return nil, 0
	}

	switch spec.kind {
	case kindFixedASCII:
		if len(data) < spec.fixed {
			return nil, 0
		}
		return data[:spec.fixed], spec.fixed

	case kindN3BCD:
		if len(data) < spec.fixed {
			return nil, 0
		}
		return data[:spec.fixed], spec.fixed

	case kindLVARASCII:
		if len(data) < 1 {
			return nil, 0
		}
		l := bcdByteToInt(data[0])
		if l < 0 || l > spec.maxLen {
			return nil, 0
		}
		if len(data) < 1+l {
			return nil, 0
		}
		return data[1 : 1+l], 1 + l

	case kindLLVARASCII:
		if len(data) < 2 {
			return nil, 0
		}
		l := bcd2BytesToInt(data[0], data[1])
		if l < 0 || l > spec.maxLen {
			return nil, 0
		}
		if len(data) < 2+l {
			return nil, 0
		}
		return data[2 : 2+l], 2 + l

	default:
		return nil, 0
	}
}

func bcdByteToInt(b byte) int {
	hi := int((b >> 4) & 0x0F)
	lo := int(b & 0x0F)
	if hi > 9 || lo > 9 {
		return -1
	}
	return hi*10 + lo
}

func bcd2BytesToInt(b1, b2 byte) int {
	d1 := int((b1 >> 4) & 0x0F)
	d2 := int(b1 & 0x0F)
	d3 := int((b2 >> 4) & 0x0F)
	d4 := int(b2 & 0x0F)

	if d1 > 9 || d2 > 9 || d3 > 9 || d4 > 9 {
		return -1
	}

	return d1*1000 + d2*100 + d3*10 + d4
}

func encodeLVARLength(n int) (byte, error) {
	if n < 0 || n > 99 {
		return 0, fmt.Errorf("lvar length out of range: %d", n)
	}
	return byte(((n/10)&0x0F)<<4 | (n%10)&0x0F), nil
}

func encodeLLVARLength(n int) ([2]byte, error) {
	var out [2]byte
	if n < 0 || n > 9999 {
		return out, fmt.Errorf("llvar length out of range: %d", n)
	}
	d1 := (n / 1000) % 10
	d2 := (n / 100) % 10
	d3 := (n / 10) % 10
	d4 := n % 10

	out[0] = byte((d1<<4)&0xF0 | (d2 & 0x0F))
	out[1] = byte((d3<<4)&0xF0 | (d4 & 0x0F))
	return out, nil
}
