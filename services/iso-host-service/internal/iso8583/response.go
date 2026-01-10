package iso8583

var echoFieldsByReqMTI = map[string]map[int]struct{}{
	"1100": {
		2: {}, 3: {}, 4: {}, 11: {}, 12: {}, 13: {}, 14: {},
		37: {}, 41: {}, 42: {}, 49: {}, 63: {},
	},
	"1420": {
		2: {}, 3: {}, 4: {}, 11: {}, 12: {}, 13: {}, 14: {},
		37: {}, 41: {}, 42: {}, 49: {},
	},
}

func NewResponse(req *Message) *Message {
	resp := &Message{
		MTI:    mapResponseMTI(req.MTI),
		Fields: make(map[int][]byte),
	}

	allowed := echoFieldsByReqMTI[req.MTI]

	for bit := range allowed {
		if val, ok := req.Fields[bit]; ok {
			resp.Fields[bit] = val
			setBit(resp.Bitmap[:], bit)
		}
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
