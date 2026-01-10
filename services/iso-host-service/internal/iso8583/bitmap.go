package iso8583

func isBitSet(bitmap []byte, bit int) bool {
	if bit < 1 || bit > 64 || len(bitmap) < 8 {
		return false
	}

	byteIndex := (bit - 1) / 8
	bitIndex := (bit - 1) % 8

	mask := byte(1 << (7 - bitIndex))
	return (bitmap[byteIndex] & mask) != 0
}

func setBit(bitmap []byte, bit int) {
	if bit < 1 || bit > 64 || len(bitmap) < 8 {
		return
	}

	byteIndex := (bit - 1) / 8
	bitIndex := (bit - 1) % 8

	mask := byte(1 << (7 - bitIndex))
	bitmap[byteIndex] |= mask
}

func clearBit(bitmap []byte, bit int) {
	if bit < 1 || bit > 64 || len(bitmap) < 8 {
		return
	}

	byteIndex := (bit - 1) / 8
	bitIndex := (bit - 1) % 8

	mask := byte(1 << (7 - bitIndex))
	bitmap[byteIndex] &^= mask
}
