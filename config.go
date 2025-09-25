package imgproxy

import (
	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	processinghandler "github.com/imgproxy/imgproxy/v3/handlers/processing"
	streamhandler "github.com/imgproxy/imgproxy/v3/handlers/stream"
	optionsparser "github.com/imgproxy/imgproxy/v3/options/parser"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/workers"
)

// HandlerConfigs holds the configurations for imgproxy handlers
type HandlerConfigs struct {
	Processing processinghandler.Config
	Stream     streamhandler.Config
}

// Config represents an instance configuration
type Config struct {
	Workers        workers.Config
	FallbackImage  auximageprovider.StaticConfig
	WatermarkImage auximageprovider.StaticConfig
	Fetcher        fetcher.Config
	Handlers       HandlerConfigs
	Server         server.Config
	Security       security.Config
	Processing     processing.Config
	OptionsParser  optionsparser.Config
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig() Config {
	return Config{
		Workers:        workers.NewDefaultConfig(),
		FallbackImage:  auximageprovider.NewDefaultStaticConfig(),
		WatermarkImage: auximageprovider.NewDefaultStaticConfig(),
		Fetcher:        fetcher.NewDefaultConfig(),
		Handlers: HandlerConfigs{
			Processing: processinghandler.NewDefaultConfig(),
			Stream:     streamhandler.NewDefaultConfig(),
		},
		Server:        server.NewDefaultConfig(),
		Security:      security.NewDefaultConfig(),
		Processing:    processing.NewDefaultConfig(),
		OptionsParser: optionsparser.NewDefaultConfig(),
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

	if _, err = workers.LoadConfigFromEnv(&c.Workers); err != nil {
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

	if _, err = security.LoadConfigFromEnv(&c.Security); err != nil {
		return nil, err
	}

	if _, err = optionsparser.LoadConfigFromEnv(&c.OptionsParser); err != nil {
		return nil, err
	}

	if _, err = processing.LoadConfigFromEnv(&c.Processing); err != nil {
		return nil, err
	}

	return c, nil
}
