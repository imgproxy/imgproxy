package imgproxy

import (
	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	processinghandler "github.com/imgproxy/imgproxy/v3/handlers/processing"
	streamhandler "github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/semaphores"
	"github.com/imgproxy/imgproxy/v3/server"
)

// HandlerConfigs holds the configurations for imgproxy handlers
type HandlerConfigs struct {
	Processing processinghandler.Config
	Stream     streamhandler.Config
}

// Config represents an instance configuration
type Config struct {
	Semaphores     semaphores.Config
	FallbackImage  auximageprovider.StaticConfig
	WatermarkImage auximageprovider.StaticConfig
	Fetcher        fetcher.Config
	Handlers       HandlerConfigs
	Server         server.Config
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig() Config {
	return Config{
		Semaphores:     semaphores.NewDefaultConfig(),
		FallbackImage:  auximageprovider.NewDefaultStaticConfig(),
		WatermarkImage: auximageprovider.NewDefaultStaticConfig(),
		Fetcher:        fetcher.NewDefaultConfig(),
		Handlers: HandlerConfigs{
			Processing: processinghandler.NewDefaultConfig(),
			Stream:     streamhandler.NewDefaultConfig(),
		},
		Server: server.NewDefaultConfig(),
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

	if _, err = semaphores.LoadConfigFromEnv(&c.Semaphores); err != nil {
		return nil, err
	}

	if _, err = fetcher.LoadConfigFromEnv(&c.Fetcher); err != nil {
		return nil, err
	}

	if _, err = processinghandler.LoadConfigFromEnv(&c.Handlers.Processing); err != nil {
		return nil, err
	}

	if _, err = streamhandler.LoadConfigFromEnv(&c.Handlers.Stream); err != nil {
		return nil, err
	}

	return c, nil
}
