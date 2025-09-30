package monitoring

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/monitoring/cloudwatch"
	"github.com/imgproxy/imgproxy/v3/monitoring/datadog"
	"github.com/imgproxy/imgproxy/v3/monitoring/newrelic"
	"github.com/imgproxy/imgproxy/v3/monitoring/otel"
	"github.com/imgproxy/imgproxy/v3/monitoring/prometheus"
)

// Config holds the configuration for all monitoring services
type Config struct {
	Prometheus    prometheus.Config // Prometheus metrics configuration
	NewRelic      newrelic.Config   // New Relic configuration
	DataDog       datadog.Config    // DataDog configuration
	OpenTelemetry otel.Config       // OpenTelemetry configuration
	CloudWatch    cloudwatch.Config // CloudWatch configuration
}

// NewDefaultConfig returns a new default configuration for all monitoring services
func NewDefaultConfig() Config {
	return Config{
		Prometheus:    prometheus.NewDefaultConfig(),
		NewRelic:      newrelic.NewDefaultConfig(),
		DataDog:       datadog.NewDefaultConfig(),
		OpenTelemetry: otel.NewDefaultConfig(),
		CloudWatch:    cloudwatch.NewDefaultConfig(),
	}
}

// LoadConfigFromEnv loads configuration for all monitoring services from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	var promErr, nrErr, ddErr, otelErr, cwErr error

	_, promErr = prometheus.LoadConfigFromEnv(&c.Prometheus)
	_, nrErr = newrelic.LoadConfigFromEnv(&c.NewRelic)
	_, ddErr = datadog.LoadConfigFromEnv(&c.DataDog)
	_, otelErr = otel.LoadConfigFromEnv(&c.OpenTelemetry)
	_, cwErr = cloudwatch.LoadConfigFromEnv(&c.CloudWatch)

	err := errors.Join(promErr, nrErr, ddErr, otelErr, cwErr)

	return c, err
}

func (c *Config) Validate() error {
	return nil
}
