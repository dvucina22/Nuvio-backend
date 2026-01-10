package iso8583

func (m *Message) SetField(bit int, value []byte) {
	if m.Fields == nil {
		m.Fields = make(map[int][]byte)
	}

	m.Fields[bit] = value
	setBit(m.Bitmap[:], bit)
}

func (m *Message) RemoveField(bit int) {
	if m.Fields != nil {
		delete(m.Fields, bit)
	}

	clearBit(m.Bitmap[:], bit)
}
