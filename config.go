package imgproxy

import (
	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/handlers"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/semaphores"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/transport"
)

// Config represents an instance configuration
type Config struct {
	HeaderWriter      headerwriter.Config
	Semaphores        semaphores.Config
	FallbackImage     auximageprovider.StaticConfig
	WatermarkImage    auximageprovider.StaticConfig
	Transport         transport.Config
	Fetcher           fetcher.Config
	ProcessingHandler handlers.Config
	StreamHandler     stream.Config
	Server            server.Config
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig() Config {
	return Config{
		HeaderWriter:      headerwriter.NewDefaultConfig(),
		Semaphores:        semaphores.NewDefaultConfig(),
		FallbackImage:     auximageprovider.NewDefaultStaticConfig(),
		WatermarkImage:    auximageprovider.NewDefaultStaticConfig(),
		Transport:         transport.NewDefaultConfig(),
		Fetcher:           fetcher.NewDefaultConfig(),
		ProcessingHandler: handlers.NewDefaultConfig(),
		StreamHandler:     stream.NewDefaultConfig(),
		Server:            server.NewDefaultConfig(),
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	var err error

	if _, err = server.LoadConfigFromEnv(&c.Server); err != nil {
		return nil, err
	}

	if _, err = auximageprovider.LoadFallbackStaticConfigFromEnv(&c.FallbackImage); err != nil {
		return nil, err
	}

	if _, err = auximageprovider.LoadWatermarkStaticConfigFromEnv(&c.WatermarkImage); err != nil {
		return nil, err
	}

	if _, err = headerwriter.LoadConfigFromEnv(&c.HeaderWriter); err != nil {
		return nil, err
	}

	if _, err = semaphores.LoadConfigFromEnv(&c.Semaphores); err != nil {
		return nil, err
	}

	if _, err = transport.LoadConfigFromEnv(&c.Transport); err != nil {
		return nil, err
	}

	if _, err = fetcher.LoadConfigFromEnv(&c.Fetcher); err != nil {
		return nil, err
	}

	if _, err = handlers.LoadConfigFromEnv(&c.ProcessingHandler); err != nil {
		return nil, err
	}

	if _, err = stream.LoadConfigFromEnv(&c.StreamHandler); err != nil {
		return nil, err
	}

	return c, nil
}
