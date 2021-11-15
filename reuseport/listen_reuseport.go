//go:build (linux || darwin) && go1.11
// +build linux darwin
// +build go1.11

package reuseport

import (
	"context"
	"net"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/imgproxy/imgproxy/v3/config"
)

func Listen(network, address string) (net.Listener, error) {
	if !config.SoReuseport {
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
