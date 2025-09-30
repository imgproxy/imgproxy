// config.go is just a shortcut for common.Config which helps to
// avoid importing of the `common` package directly.
package transport

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/azure"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/fs"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/gcs"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/generichttp"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/s3"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/swift"
)

var (
	IMGPROXY_USE_ABS   = env.Describe("IMGPROXY_USE_ABS", "boolean")
	IMGPROXY_USE_GCS   = env.Describe("IMGPROXY_GCS_ENABLED", "boolean")
	IMGPROXY_USE_S3    = env.Describe("IMGPROXY_USE_S3", "boolean")
	IMGPROXY_USE_SWIFT = env.Describe("IMGPROXY_USE_SWIFT", "boolean")
)

// Config represents configuration of the transport package
type Config struct {
	HTTP generichttp.Config

	Local fs.Config

	ABSEnabled bool
	ABS        azure.Config

	GCSEnabled bool
	GCS        gcs.Config

	S3Enabled bool
	S3        s3.Config

	SwiftEnabled bool
	Swift        swift.Config
}

// NewDefaultConfig returns a new default transport configuration
func NewDefaultConfig() Config {
	return Config{
		HTTP:         generichttp.NewDefaultConfig(),
		Local:        fs.NewDefaultConfig(),
		ABSEnabled:   false,
		ABS:          azure.NewDefaultConfig(),
		GCSEnabled:   false,
		GCS:          gcs.NewDefaultConfig(),
		S3Enabled:    false,
		S3:           s3.NewDefaultConfig(),
		SwiftEnabled: false,
		Swift:        swift.NewDefaultConfig(),
	}
}

// LoadConfigFromEnv loads transport configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	_, genericErr := generichttp.LoadConfigFromEnv(&c.HTTP)
	_, localErr := fs.LoadConfigFromEnv(&c.Local)
	_, azureErr := azure.LoadConfigFromEnv(&c.ABS)
	_, gcsErr := gcs.LoadConfigFromEnv(&c.GCS)
	_, s3Err := s3.LoadConfigFromEnv(&c.S3)
	_, swiftErr := swift.LoadConfigFromEnv(&c.Swift)

	err := errors.Join(
		genericErr,
		localErr,
		azureErr,
		gcsErr,
		s3Err,
		swiftErr,
		env.Bool(&c.ABSEnabled, IMGPROXY_USE_ABS),
		env.Bool(&c.GCSEnabled, IMGPROXY_USE_GCS),
		env.Bool(&c.S3Enabled, IMGPROXY_USE_S3),
		env.Bool(&c.SwiftEnabled, IMGPROXY_USE_SWIFT),
	)

	return c, err
}

func (c *Config) Validate() error {
	// We won't validate upstream config here: they might not be used
	return nil
}
