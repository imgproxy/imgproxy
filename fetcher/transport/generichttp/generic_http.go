// Package generichttp provides Generic HTTP transport for imgproxy
package generichttp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/imgproxy/imgproxy/v4/privatenet"
)

func New(verifyNetworks bool, config *Config) (*http.Transport, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	if verifyNetworks {
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			return VerifySourceNetwork(address, config)
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
			MaxReadFrameSize:          16 * 1024,
			PingTimeout:               5 * time.Second,
			SendPingTimeout:           time.Second,
		},
	}

	if config.ClientKeepAliveTimeout <= 0 {
		transport.MaxIdleConnsPerHost = -1
		transport.DisableKeepAlives = true
	}

	if config.IgnoreSslVerification {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return transport, nil
}

func VerifySourceNetwork(addr string, config *Config) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return newSourceAddressError(fmt.Sprintf("Invalid source address: %s", addr))
	}

	if !config.AllowLoopbackSourceAddresses && privatenet.IsLoopback(ip) {
		return newSourceAddressError(fmt.Sprintf("Loopback source address is not allowed: %s", addr))
	}

	if !config.AllowLinkLocalSourceAddresses && privatenet.IsLinkLocal(ip) {
		return newSourceAddressError(fmt.Sprintf("Link-local source address is not allowed: %s", addr))
	}

	if !config.AllowPrivateSourceAddresses && privatenet.IsPrivate(ip) {
		return newSourceAddressError(fmt.Sprintf("Private source address is not allowed: %s", addr))
	}

	return nil
}
