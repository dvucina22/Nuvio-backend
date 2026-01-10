package iso8583

import (
	"fmt"
)

func Decode(raw []byte) (*Message, error) {
	if len(raw) < 2 {
		return nil, fmt.Errorf("invalid message: missing 2-byte length header")
	}

	length := int(raw[0])<<8 | int(raw[1])

	if length <= 0 {
		return nil, fmt.Errorf("invalid message: length=%d", length)
	}

	if length > 1024 {
		return nil, fmt.Errorf("invalid message: length=%d exceeds max 1024", length)
	}

	if len(raw) < 2+length {
		return nil, fmt.Errorf("invalid message: header length=%d but have=%d bytes", length, len(raw)-2)
	}

	payload := raw[2 : 2+length]

	if len(payload) < 10 {
		return nil, fmt.Errorf("invalid message: payload too short for MTI+bitmap (%d)", len(payload))
	}

	mti, err := decodeMTI(payload[:2])
	if err != nil {
		return nil, err
	}

	bitmapBytes := payload[2:10]

	fields := make(map[int][]byte)
	offset := 10

	for i := 1; i <= 64; i++ {
		if !isBitSet(bitmapBytes, i) {
			continue
		}

		if offset >= len(payload) {
			return nil, fmt.Errorf("invalid message: field %d set but payload ended", i)
		}

		data, size := decodeField(i, payload[offset:])
		if size <= 0 {
			return nil, fmt.Errorf("invalid message: failed to decode field %d at offset %d", i, offset)
		}

		if offset+size > len(payload) {
			return nil, fmt.Errorf("invalid message: field %d overruns payload (offset=%d size=%d len=%d)", i, offset, size, len(payload))
		}

		fields[i] = data
		offset += size
	}

	var bitmap [8]byte
	copy(bitmap[:], bitmapBytes)

	return &Message{
		MTI:    mti,
		Bitmap: bitmap,
		Fields: fields,
	}, nil
}

func decodeMTI(b []byte) (string, error) {
	if len(b) != 2 {
		return "", fmt.Errorf("invalid MTI length: %d", len(b))
	}

	d1 := int((b[0] >> 4) & 0x0F)
	d2 := int(b[0] & 0x0F)
	d3 := int((b[1] >> 4) & 0x0F)
	d4 := int(b[1] & 0x0F)

	if d1 > 9 || d2 > 9 || d3 > 9 || d4 > 9 {
		return "", fmt.Errorf("invalid MTI BCD: %02X%02X", b[0], b[1])
	}

	return fmt.Sprintf("%d%d%d%d", d1, d2, d3, d4), nil
}
