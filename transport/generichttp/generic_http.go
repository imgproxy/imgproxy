// Generic HTTP transport for imgproxy
package generichttp

import (
	"crypto/tls"
	"net"
	"net/http"
	"syscall"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/security"
	"golang.org/x/net/http2"
)

func New(verifyNetworks bool) (*http.Transport, error) {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	if verifyNetworks {
		dialer.Control = func(network, address string, c syscall.RawConn) error {
			return security.VerifySourceNetwork(address)
		}
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   config.Workers + 1,
		IdleConnTimeout:       time.Duration(config.ClientKeepAliveTimeout) * time.Second,
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
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
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
