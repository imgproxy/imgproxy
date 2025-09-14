//go:build (!linux && !darwin) || !go1.11
// +build !linux,!darwin !go1.11

package reuseport

import (
	"log/slog"
	"net"
)

func Listen(network, address string, reuse bool) (net.Listener, error) {
	if reuse {
		slog.Warn("SO_REUSEPORT support is not implemented for your OS or Go version")
	}

	return net.Listen(network, address)
}
