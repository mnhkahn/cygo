package goget

import (
	"net"
	"time"
)

func KeepAliveDialTimeout(network, addr string) (net.Conn, error) {
	dial := net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// conn, err := net.DialTimeout(network, addr, 1*time.Second)
	conn, err := dial.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	return conn, err
}
