// +build !linux,!darwin !go1.11

package reuseport

import (
	"net"
)

func Listen(network, address string) (net.Listener, error) {
	if config.SoReuseport {
		log.Warning("SO_REUSEPORT support is not implemented for your OS or Go version")
	}

	return net.Listen(network, address)
}
