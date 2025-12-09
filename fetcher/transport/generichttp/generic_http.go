// Package generichttp provides Generic HTTP transport for imgproxy
package generichttp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"

	"golang.org/x/net/http2"
)

func New(verifyNetworks bool, config *Config) (*http.Transport, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	if verifyNetworks {
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			return verifySourceNetwork(address, config)
		}
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   100,
		IdleConnTimeout:       config.ClientKeepAliveTimeout,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     false,
		DisableCompression:    true,

		HTTP2: &http.HTTP2Config{
			MaxReceiveBufferPerStream: 128 * 1024,
		},
	}

	if config.ClientKeepAliveTimeout <= 0 {
		transport.MaxIdleConnsPerHost = -1
		transport.DisableKeepAlives = true
	}

	if config.IgnoreSslVerification {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	transport2, err := http2.ConfigureTransports(transport)
	if err != nil {
		return nil, err
	}

	// TODO: Move this to transport.HTTP2 when https://go.dev/issue/67813 is closed
	transport2.MaxReadFrameSize = 16 * 1024
	transport2.PingTimeout = 5 * time.Second
	transport2.ReadIdleTimeout = time.Second

	return transport, nil
}

func verifySourceNetwork(addr string, config *Config) error {
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
