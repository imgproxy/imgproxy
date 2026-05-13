package imgproxy

import (
	"github.com/imgproxy/imgproxy/v4/auximageprovider"
	"github.com/imgproxy/imgproxy/v4/clientfeatures"
	"github.com/imgproxy/imgproxy/v4/cookies"
	"github.com/imgproxy/imgproxy/v4/ensure"
	"github.com/imgproxy/imgproxy/v4/errorreport"
	"github.com/imgproxy/imgproxy/v4/fetcher"
	processinghandler "github.com/imgproxy/imgproxy/v4/handlers/processing"
	streamhandler "github.com/imgproxy/imgproxy/v4/handlers/stream"
	"github.com/imgproxy/imgproxy/v4/httpheaders/conditionalheaders"
	"github.com/imgproxy/imgproxy/v4/monitoring"
	"github.com/imgproxy/imgproxy/v4/monitoring/prometheus"
	optionsparser "github.com/imgproxy/imgproxy/v4/options/parser"
	"github.com/imgproxy/imgproxy/v4/processing"
	"github.com/imgproxy/imgproxy/v4/security"
	"github.com/imgproxy/imgproxy/v4/server"
	"github.com/imgproxy/imgproxy/v4/workers"
)

// HandlerConfigs holds the configurations for imgproxy handlers
type HandlerConfigs struct {
	Processing processinghandler.Config
	Stream     streamhandler.Config
}

// Config represents an instance configuration
type Config struct {
	Workers            workers.Config
	FallbackImage      auximageprovider.StaticConfig
	WatermarkImage     auximageprovider.StaticConfig
	Fetcher            fetcher.Config
	ClientFeatures     clientfeatures.Config
	Handlers           HandlerConfigs
	Server             server.Config
	Security           security.Config
	Processing         processing.Config
	OptionsParser      optionsparser.Config
	Cookies            cookies.Config
	Monitoring         monitoring.Config
	ErrorReport        errorreport.Config
	ConditionalHeaders conditionalheaders.Config
}

// NewDefaultConfig creates a new default configuration
func NewDefaultConfig() Config {
	return Config{
		Workers:        workers.NewDefaultConfig(),
		FallbackImage:  auximageprovider.NewDefaultStaticConfig(),
		WatermarkImage: auximageprovider.NewDefaultStaticConfig(),
		Fetcher:        fetcher.NewDefaultConfig(),
		ClientFeatures: clientfeatures.NewDefaultConfig(),
		Handlers: HandlerConfigs{
			Processing: processinghandler.NewDefaultConfig(),
			Stream:     streamhandler.NewDefaultConfig(),
		},
		Server:             server.NewDefaultConfig(),
		Security:           security.NewDefaultConfig(),
		Processing:         processing.NewDefaultConfig(),
		OptionsParser:      optionsparser.NewDefaultConfig(),
		Cookies:            cookies.NewDefaultConfig(),
		Monitoring:         monitoring.NewDefaultConfig(),
		ErrorReport:        errorreport.NewDefaultConfig(),
		ConditionalHeaders: conditionalheaders.NewDefaultConfig(),
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

	if _, err = clientfeatures.LoadConfigFromEnv(&c.ClientFeatures); err != nil {
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

	if _, err = cookies.LoadConfigFromEnv(&c.Cookies); err != nil {
		return nil, err
	}

	if _, err = monitoring.LoadConfigFromEnv(&c.Monitoring); err != nil {
		return nil, err
	}

	if _, err = errorreport.LoadConfigFromEnv(&c.ErrorReport); err != nil {
		return nil, err
	}

	if _, err = conditionalheaders.LoadConfigFromEnv(&c.ConditionalHeaders); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) Validate() error {
	if c.Monitoring.Prometheus.Enabled() && c.Monitoring.Prometheus.Bind == c.Server.Bind {
		return prometheus.IMGPROXY_PROMETHEUS_BIND.Errorf("should be different than IMGPROXY_BIND: %s", c.Server.Bind)
	}

	return nil
}
