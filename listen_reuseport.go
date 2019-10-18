// +build linux darwin
// +build go1.11

package main

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

func listenReuseport(network, address string) (net.Listener, error) {
	if !conf.SoReuseport {
		return net.Listen(network, address)
	}

	lc := net.ListenConfig{
		Control: func(_, _ string, c syscall.RawConn) error {
			var cerr error
			err := c.Control(func(fd uintptr) {
				cerr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			})
			if err != nil {
				return err
			}
			return cerr
		},
	}

	return lc.Listen(context.Background(), network, address)
}
