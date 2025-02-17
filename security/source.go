package security

import (
	"fmt"
	"net"

	"github.com/imgproxy/imgproxy/v3/config"
)

func VerifySourceURL(imageURL string) error {
	if len(config.AllowedSources) == 0 {
		return nil
	}

	for _, allowedSource := range config.AllowedSources {
		if allowedSource.MatchString(imageURL) {
			return nil
		}
	}

	return newSourceURLError(imageURL)
}

func VerifySourceNetwork(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return newSourceAddressError(fmt.Sprintf("Invalid source address: %s", addr))
	}

	if !config.AllowLoopbackSourceAddresses && (ip.IsLoopback() || ip.IsUnspecified()) {
		return newSourceAddressError(fmt.Sprintf("Loopback source address is not allowed: %s", addr))
	}

	if !config.AllowLinkLocalSourceAddresses && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
		return newSourceAddressError(fmt.Sprintf("Link-local source address is not allowed: %s", addr))
	}

	if !config.AllowPrivateSourceAddresses && ip.IsPrivate() {
		return newSourceAddressError(fmt.Sprintf("Private source address is not allowed: %s", addr))
	}

	return nil
}
