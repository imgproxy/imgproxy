// config.go is just a shortcut for common.Config which helps to
// avoid importing of the `common` package directly.
package transport

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/transport/azure"
	"github.com/imgproxy/imgproxy/v3/transport/fs"
	"github.com/imgproxy/imgproxy/v3/transport/gcs"
	"github.com/imgproxy/imgproxy/v3/transport/generichttp"
	"github.com/imgproxy/imgproxy/v3/transport/s3"
	"github.com/imgproxy/imgproxy/v3/transport/swift"
)

// Config represents configuration of the transport package
type Config struct {
	HTTP *generichttp.Config

	LocalEnabled bool
	Local        *fs.Config

	ABSEnabled bool
	ABS        *azure.Config

	GCSEnabled bool
	GCS        *gcs.Config

	S3Enabled bool
	S3        *s3.Config

	SwiftEnabled bool
	Swift        *swift.Config
}

// NewDefaultConfig returns a new default transport configuration
func NewDefaultConfig() *Config {
	return &Config{
		HTTP:         generichttp.NewDefaultConfig(),
		LocalEnabled: false,
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

// LoadFromEnv loads transport configuration from environment variables
func LoadFromEnv(c *Config) (*Config, error) {
	var err error

	if c.HTTP, err = generichttp.LoadFromEnv(c.HTTP); err != nil {
		return nil, err
	}

	if c.Local, err = fs.LoadFromEnv(c.Local); err != nil {
		return nil, err
	}

	if c.ABS, err = azure.LoadFromEnv(c.ABS); err != nil {
		return nil, err
	}

	if c.GCS, err = gcs.LoadFromEnv(c.GCS); err != nil {
		return nil, err
	}

	if c.S3, err = s3.LoadFromEnv(c.S3); err != nil {
		return nil, err
	}

	if c.Swift, err = swift.LoadFromEnv(c.Swift); err != nil {
		return nil, err
	}

	c.LocalEnabled = config.LocalFileSystemRoot != ""
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
