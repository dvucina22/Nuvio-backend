package iso8583

func FormatErrorResponse(reqMTI string) *Message {
	resp := &Message{
		MTI:    mapResponseMTI(reqMTI),
		Fields: make(map[int][]byte),
	}
	resp.SetField(39, []byte("30"))
	return resp
}
