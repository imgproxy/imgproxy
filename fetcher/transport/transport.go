// Package transport provides a custom HTTP transport that supports multiple protocols
// such as S3, GCS, ABS, Swift, and local file system.
package transport

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"

	absStorage "github.com/imgproxy/imgproxy/v3/storage/abs"
	fsStorage "github.com/imgproxy/imgproxy/v3/storage/fs"
	gcsStorage "github.com/imgproxy/imgproxy/v3/storage/gcs"
	s3Storage "github.com/imgproxy/imgproxy/v3/storage/s3"
	swiftStorage "github.com/imgproxy/imgproxy/v3/storage/swift"
)

// Transport is a wrapper around http.Transport which allows to track registered protocols
type Transport struct {
	config    *Config
	transport *http.Transport
	schemes   map[string]struct{}
}

// New creates a new HTTP transport with no protocols registered
func New(config *Config) (*Transport, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	transport, err := generichttp.New(true, &config.HTTP)
	if err != nil {
		return nil, err
	}

	// http and https are always registered
	schemes := map[string]struct{}{
		"http":  {},
		"https": {},
	}

	t := &Transport{
		config:    config,
		transport: transport,
		schemes:   schemes,
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
	slog.Info("Scheme registered", "scheme", scheme)
}

// IsProtocolRegistered checks if a protocol is registered in the transport
func (t *Transport) IsProtocolRegistered(scheme string) bool {
	_, ok := t.schemes[scheme]
	return ok
}

// RegisterAllProtocols registers all enabled protocols in the given transport
func (t *Transport) registerAllProtocols() error {
	sep := t.config.SourceURLQuerySeparator // shortcut

	transp, err := generichttp.New(false, &t.config.HTTP)
	if err != nil {
		return err
	}

	if t.config.Local.Root != "" {
		tr, err := fsStorage.New(&t.config.Local)
		if err != nil {
			return err
		}
		t.RegisterProtocol("local", NewRoundTripper(tr, sep))
	}

	if t.config.S3Enabled {
		tr, err := s3Storage.New(&t.config.S3, transp)
		if err != nil {
			return err
		}
		t.RegisterProtocol("s3", NewRoundTripper(tr, sep))
	}

	if t.config.GCSEnabled {
		tr, err := gcsStorage.New(&t.config.GCS, transp, true)
		if err != nil {
			return err
		}
		t.RegisterProtocol("gs", NewRoundTripper(tr, sep))
	}

	if t.config.ABSEnabled {
		tr, err := absStorage.New(&t.config.ABS, transp)
		if err != nil {
			return err
		}
		t.RegisterProtocol("abs", NewRoundTripper(tr, sep))
	}

	if t.config.SwiftEnabled {
		tr, err := swiftStorage.New(
			context.Background(),
			&t.config.Swift,
			transp,
		)
		if err != nil {
			return err
		}

		t.RegisterProtocol("swift", NewRoundTripper(tr, sep))
	}

	return nil
}
