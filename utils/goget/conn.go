package goget

import (
	"net"
	"time"
)

func PrintLocalDial(network, addr string) (net.Conn, error) {
	// dial := net.Dialer{
	// 	Timeout: 1 * time.Second,
	// 	// KeepAlive: 30 * time.Second,
	// 	Deadline: time.Now().Add(1 * time.Second),
	// }

	// conn, err := dial.Dial(network, addr)
	// if err != nil {
	// 	return conn, err
	// }
	conn, err := net.DialTimeout(network, addr, 1*time.Second)
	if err != nil {
		return nil, err
	}
	conn.SetDeadline(time.Now().Add(10 * time.Second))

	return conn, err
}
