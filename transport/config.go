// config.go is just a shortcut for common.Config which helps to
// avoid importing of the `common` package directly.
package transport

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/transport/azure"
	"github.com/imgproxy/imgproxy/v3/transport/fs"
	"github.com/imgproxy/imgproxy/v3/transport/gcs"
	"github.com/imgproxy/imgproxy/v3/transport/generichttp"
	"github.com/imgproxy/imgproxy/v3/transport/s3"
	"github.com/imgproxy/imgproxy/v3/transport/swift"
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

	var err error

	if _, err = generichttp.LoadConfigFromEnv(&c.HTTP); err != nil {
		return nil, err
	}

	if _, err = fs.LoadConfigFromEnv(&c.Local); err != nil {
		return nil, err
	}

	if _, err = azure.LoadConfigFromEnv(&c.ABS); err != nil {
		return nil, err
	}

	if _, err = gcs.LoadConfigFromEnv(&c.GCS); err != nil {
		return nil, err
	}

	if _, err = s3.LoadConfigFromEnv(&c.S3); err != nil {
		return nil, err
	}

	if _, err = swift.LoadConfigFromEnv(&c.Swift); err != nil {
		return nil, err
	}

	c.ABSEnabled = config.ABSEnabled
	c.GCSEnabled = config.GCSEnabled
	c.S3Enabled = config.S3Enabled
	c.SwiftEnabled = config.SwiftEnabled

	return c, nil
}

func (c *Config) Validate() error {
	// We won't validate upstream config here: they might not be used
	return nil
}
