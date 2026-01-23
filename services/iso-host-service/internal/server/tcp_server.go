package server

import (
	"log"
	"net"
)

func Start(addr string, handler func([]byte) ([]byte, error)) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Println("Host listening on", addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			buf := make([]byte, 1024)
			n, err := c.Read(buf)
			if err != nil {
				return
			}

			resp, err := handler(buf[:n])
			if err != nil {
				return
			}

			log.Printf("ISO SEND: addr=%s \npayload_len=%d \nhex=%x", c.RemoteAddr(), len(resp), resp)

			c.Write(resp)
		}(conn)
	}
}
