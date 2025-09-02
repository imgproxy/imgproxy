package imgproxy

import (
	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/loadenv"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	processinghandler "github.com/imgproxy/imgproxy/v3/handlers/processing"
	"github.com/imgproxy/imgproxy/v3/handlers/stream"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/semaphores"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/transport"
)

// Config represents an instance configuration
type Config struct {
	HeaderWriter      *headerwriter.Config
	Semaphores        *semaphores.Config
	FallbackImage     *auximageprovider.StaticConfig
	Transport         *transport.Config
	Fetcher           *fetcher.Config
	ProcessingHandler *processinghandler.Config
	StreamHandler     *stream.Config
	Server            *server.Config
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig() *Config {
	return &Config{
		HeaderWriter:      headerwriter.NewDefaultConfig(),
		Semaphores:        semaphores.NewDefaultConfig(),
		FallbackImage:     auximageprovider.NewDefaultStaticConfig(),
		Transport:         transport.NewDefaultConfig(),
		Fetcher:           fetcher.NewDefaultConfig(),
		ProcessingHandler: processinghandler.NewDefaultConfig(),
		StreamHandler:     stream.NewDefaultConfig(),
		Server:            server.NewDefaultConfig(),
	}
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv(c *Config) (*Config, error) {
	// NOTE: This is temporary workaround. We have to load env vars in config.go before
	// actually configuring ImgProxy instance because for now we use it as a source of truth.
	// Will be removed once we move env var loading to imgproxy.go
	if err := loadenv.Load(); err != nil {
		return nil, err
	}

	if err := config.Configure(); err != nil {
		return nil, err
	}
	// NOTE: End of temporary workaround.

	var err error

	if c.Server, err = server.LoadFromEnv(c.Server); err != nil {
		return nil, err
	}

	if c.FallbackImage, err = auximageprovider.LoadFallbackStaticConfigFromEnv(c.FallbackImage); err != nil {
		return nil, err
	}

	if c.HeaderWriter, err = headerwriter.LoadFromEnv(c.HeaderWriter); err != nil {
		return nil, err
	}

	if c.Semaphores, err = semaphores.LoadFromEnv(c.Semaphores); err != nil {
		return nil, err
	}

	if c.Transport, err = transport.LoadFromEnv(c.Transport); err != nil {
		return nil, err
	}

	if c.Fetcher, err = fetcher.LoadFromEnv(c.Fetcher); err != nil {
		return nil, err
	}

	if c.ProcessingHandler, err = processinghandler.LoadFromEnv(c.ProcessingHandler); err != nil {
		return nil, err
	}

	if c.StreamHandler, err = stream.LoadFromEnv(c.StreamHandler); err != nil {
		return nil, err
	}

	return c, nil
}
