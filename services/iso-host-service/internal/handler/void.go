package handler

import (
	"github.com/iso-host-service/internal/iso8583"
)

func HandleVoid(msg *iso8583.Message) (*iso8583.Message, error) {
	resp := iso8583.NewResponse(msg)

	resp.SetField(39, []byte("00"))
	resp.RemoveField(38)

	return resp, nil
}
