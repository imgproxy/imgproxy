package security

import (
	"errors"
	"fmt"
	"net"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

var ErrSourceAddressNotAllowed = errors.New("source address is not allowed")
var ErrInvalidSourceAddress = errors.New("invalid source address")

func VerifySourceURL(imageURL string) error {
	if len(config.AllowedSources) == 0 {
		return nil
	}

	for _, allowedSource := range config.AllowedSources {
		if allowedSource.MatchString(imageURL) {
			return nil
		}
	}

	return ierrors.New(
		404,
		fmt.Sprintf("Source URL is not allowed: %s", imageURL),
		"Invalid source",
	)
}

func VerifySourceNetwork(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return ErrInvalidSourceAddress
	}

	if !config.AllowLoopbackSourceAddresses && ip.IsLoopback() {
		return ErrSourceAddressNotAllowed
	}

	if !config.AllowLinkLocalSourceAddresses && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
		return ErrSourceAddressNotAllowed
	}

	if !config.AllowPrivateSourceAddresses && ip.IsPrivate() {
		return ErrSourceAddressNotAllowed
	}

	return nil
}
