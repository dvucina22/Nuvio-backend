package iso8583

func Encode(msg *Message) []byte {
	payload := []byte{}

	payload = append(payload, encodeBCD(msg.MTI)...)

	bitmap := buildBitmapFromFields(msg.Fields)
	msg.Bitmap = bitmap
	payload = append(payload, msg.Bitmap[:]...)

	for i := 1; i <= 64; i++ {
		if val, ok := msg.Fields[i]; ok {
			payload = append(payload, encodeField(i, val)...)
		}
	}

	length := len(payload)
	header := []byte{byte(length >> 8), byte(length)}

	return append(header, payload...)
}

func buildBitmapFromFields(fields map[int][]byte) [8]byte {
	var bitmap [8]byte

	for bit := range fields {
		if bit < 1 || bit > 64 {
			continue
		}
		setBit(bitmap[:], bit)
	}

	return bitmap
}

func encodeField(bit int, value []byte) []byte {
	spec, ok := specs[bit]
	if !ok {
		return value
	}

	switch spec.kind {
	case kindFixedASCII:
		if len(value) == spec.fixed {
			return value
		}

		out := make([]byte, spec.fixed)

		if len(value) > spec.fixed {
			copy(out, value[:spec.fixed])
			return out
		}

		pad := spec.fixed - len(value)
		for i := 0; i < pad; i++ {
			out[i] = '0'
		}
		copy(out[pad:], value)
		return out

	case kindN3BCD:
		if len(value) == 2 {
			return value
		}

		if len(value) == 3 && isAllDigits(value) {
			d1 := value[0] - '0'
			d2 := value[1] - '0'
			d3 := value[2] - '0'

			return []byte{
				byte(0x00 | (d1 & 0x0F)),
				byte(((d2 & 0x0F) << 4) | (d3 & 0x0F)),
			}
		}

		return value

	case kindLVARASCII:
		l := len(value)
		if l > spec.maxLen {
			value = value[:spec.maxLen]
			l = len(value)
		}

		prefix, err := encodeLVARLength(l)
		if err != nil {
			return value
		}

		out := make([]byte, 0, 1+l)
		out = append(out, prefix)
		out = append(out, value...)
		return out

	case kindLLVARASCII:
		l := len(value)
		if l > spec.maxLen {
			value = value[:spec.maxLen]
			l = len(value)
		}

		prefix, err := encodeLLVARLength(l)
		if err != nil {
			return value
		}

		out := make([]byte, 0, 2+l)
		out = append(out, prefix[0], prefix[1])
		out = append(out, value...)
		return out

	default:
		return value
	}
}

func encodeBCD(mti string) []byte {
	if len(mti) != 4 {
		return []byte{0x00, 0x00}
	}

	d1 := mti[0] - '0'
	d2 := mti[1] - '0'
	d3 := mti[2] - '0'
	d4 := mti[3] - '0'

	if d1 > 9 || d2 > 9 || d3 > 9 || d4 > 9 {
		return []byte{0x00, 0x00}
	}

	return []byte{
		byte((d1 << 4) | d2),
		byte((d3 << 4) | d4),
	}
}

func isAllDigits(b []byte) bool {
	for _, c := range b {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
