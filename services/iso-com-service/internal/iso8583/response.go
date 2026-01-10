package iso8583

func NewResponse(req *Message) *Message {
	resp := &Message{
		MTI:    mapResponseMTI(req.MTI),
		Fields: make(map[int][]byte, len(req.Fields)),
	}

	for i := 0; i < 8; i++ {
		resp.Bitmap[i] = 0
	}

	for bit, val := range req.Fields {
		resp.Fields[bit] = val
		setBit(resp.Bitmap[:], bit)
	}

	return resp
}

func mapResponseMTI(reqMTI string) string {
	switch reqMTI {
	case "1100":
		return "1110"
	case "1420":
		return "1430"
	default:
		return reqMTI
	}
}
