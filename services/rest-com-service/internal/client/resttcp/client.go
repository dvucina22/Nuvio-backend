package resttcp

import (
	"context"
	"io"
	"net"
	"time"
)

type Client struct {
	addr string
}

func New(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) Send(ctx context.Context, payload []byte) ([]byte, error) {
	d := net.Dialer{Timeout: 3 * time.Second}

	conn, err := d.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(6 * time.Second))

	if _, err := conn.Write(payload); err != nil {
		return nil, err
	}

	resp, err := io.ReadAll(conn)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
