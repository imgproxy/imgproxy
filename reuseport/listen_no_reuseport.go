//go:build (!linux && !darwin) || !go1.11
// +build !linux,!darwin !go1.11

package reuseport

import (
	"net"

	log "github.com/sirupsen/logrus"
)

func Listen(network, address string, reuse bool) (net.Listener, error) {
	if reuse {
		log.Warning("SO_REUSEPORT support is not implemented for your OS or Go version")
	}

	return net.Listen(network, address)
}
