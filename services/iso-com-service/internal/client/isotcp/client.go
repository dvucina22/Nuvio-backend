package isotcp

import (
	"context"
	"fmt"
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

	hdr := make([]byte, 2)
	if _, err := readExact(conn, hdr); err != nil {
		return nil, err
	}

	n := int(hdr[0])<<8 | int(hdr[1])
	if n <= 0 || n > 4096 {
		return nil, fmt.Errorf("invalid response length: %d", n)
	}

	body := make([]byte, n)
	if _, err := readExact(conn, body); err != nil {
		return nil, err
	}

	return append(hdr, body...), nil
}

func readExact(conn net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		if err != nil {
			return total, err
		}
		total += n
	}
	return total, nil
}
