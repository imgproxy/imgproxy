// +build !linux,!darwin !go1.11

package main

import (
	"net"
)

func listenReuseport(network, address string) (net.Listener, error) {
	if conf.SoReuseport {
		logWarning("SO_REUSEPORT support is not implemented for your OS or Go version")
	}

	return net.Listen(network, address)
}
