// Package transport provides a custom HTTP transport that supports multiple protocols
// such as S3, GCS, ABS, Swift, and local file system.
package transport

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v3/transport/generichttp"

	azureTransport "github.com/imgproxy/imgproxy/v3/transport/azure"
	fsTransport "github.com/imgproxy/imgproxy/v3/transport/fs"
	gcsTransport "github.com/imgproxy/imgproxy/v3/transport/gcs"
	s3Transport "github.com/imgproxy/imgproxy/v3/transport/s3"
	swiftTransport "github.com/imgproxy/imgproxy/v3/transport/swift"
)

// Transport is a wrapper around http.Transport which allows to track registered protocols
type Transport struct {
	config            *Config
	transport         *http.Transport
	upstreamTransport *http.Transport
	schemes           map[string]struct{}
}

// New creates a new HTTP transport with no protocols registered
func New(config *Config) (*Transport, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	transport, err := generichttp.New(true, config.HTTP)
	if err != nil {
		return nil, err
	}

	upstreamTransport, err := generichttp.New(false, config.HTTP)
	if err != nil {
		return nil, err
	}

	// http and https are always registered
	schemes := map[string]struct{}{
		"http":  {},
		"https": {},
	}

	t := &Transport{
		config:            config,
		transport:         transport,
		upstreamTransport: upstreamTransport,
		schemes:           schemes,
	}

	err = t.registerAllProtocols()
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Transport returns the underlying http.Transport
func (t *Transport) Transport() *http.Transport {
	return t.transport
}

// RegisterProtocol registers a new transport protocol with the transport
func (t *Transport) RegisterProtocol(scheme string, rt http.RoundTripper) {
	t.transport.RegisterProtocol(scheme, rt)
	t.schemes[scheme] = struct{}{}
}

// IsProtocolRegistered checks if a protocol is registered in the transport
func (t *Transport) IsProtocolRegistered(scheme string) bool {
	_, ok := t.schemes[scheme]
	return ok
}

// RegisterAllProtocols registers all enabled protocols in the given transport
func (t *Transport) registerAllProtocols() error {
	if t.config.Local.Root != "" {
		t.RegisterProtocol("local", fsTransport.New(t.config.Local))
	}

	if t.config.S3Enabled {
		if tr, err := s3Transport.New(t.config.S3, t.upstreamTransport); err != nil {
			return err
		} else {
			t.RegisterProtocol("s3", tr)
		}
	}

	if t.config.GCSEnabled {
		if tr, err := gcsTransport.New(t.config.GCS, t.upstreamTransport); err != nil {
			return err
		} else {
			t.RegisterProtocol("gs", tr)
		}
	}

	if t.config.ABSEnabled {
		if tr, err := azureTransport.New(t.config.ABS, t.upstreamTransport); err != nil {
			return err
		} else {
			t.RegisterProtocol("abs", tr)
		}
	}

	if t.config.SwiftEnabled {
		if tr, err := swiftTransport.New(t.config.Swift, t.upstreamTransport); err != nil {
			return err
		} else {
			t.RegisterProtocol("swift", tr)
		}
	}

	return nil
}
