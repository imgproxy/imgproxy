package datadog

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

const (
	ddDocsUrl = "https://docs.datadoghq.com/tracing/trace_collection/library_config/go"
)

var (
	IMGPROXY_DATADOG_ENABLE                    = env.Bool("IMGPROXY_DATADOG_ENABLE")
	IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS = env.Bool("IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS")
	IMGPROXY_DATADOG_PROPAGATE_EXTERNAL        = env.Bool("IMGPROXY_DATADOG_PROPAGATE_EXTERNAL")

	DD_SERVICE            = env.String("DD_SERVICE").WithDocsURL(ddDocsUrl)
	DD_TRACE_STARTUP_LOGS = env.Bool("DD_TRACE_STARTUP_LOGS").WithDocsURL(ddDocsUrl)
	DD_AGENT_HOST         = env.String("DD_AGENT_HOST").WithDocsURL(ddDocsUrl)
	DD_TRACE_AGENT_PORT   = env.Int("DD_TRACE_AGENT_PORT").WithDocsURL(ddDocsUrl)
	DD_DOGSTATSD_PORT     = env.Int("DD_DOGSTATSD_PORT").WithDocsURL(ddDocsUrl)
)

// Config holds the configuration for DataDog monitoring
type Config struct {
	Enable           bool          // Enable DataDog tracing
	EnableMetrics    bool          // Enable DataDog metrics collection
	PropagateExt     bool          // Enable propagation of tracing headers for external services
	Service          string        // DataDog service name
	TraceStartupLogs bool          // Enable trace startup logs
	AgentHost        string        // DataDog agent host
	TracePort        int           // DataDog tracer port
	StatsDPort       int           // DataDog StatsD port
	MetricsInterval  time.Duration // Interval for sending metrics to DataDog
}

// NewDefaultConfig returns a new default configuration for DataDog monitoring
func NewDefaultConfig() Config {
	return Config{
		Enable:           false,
		EnableMetrics:    false,
		PropagateExt:     false,
		Service:          "imgproxy",
		TraceStartupLogs: false,
		AgentHost:        "localhost",
		TracePort:        8126,
		StatsDPort:       8125,
		MetricsInterval:  10 * time.Second,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_DATADOG_ENABLE.Parse(&c.Enable),
		IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS.Parse(&c.EnableMetrics),
		IMGPROXY_DATADOG_PROPAGATE_EXTERNAL.Parse(&c.PropagateExt),
		DD_SERVICE.Parse(&c.Service),
		DD_TRACE_STARTUP_LOGS.Parse(&c.TraceStartupLogs),
		DD_AGENT_HOST.Parse(&c.AgentHost),
		DD_TRACE_AGENT_PORT.Parse(&c.TracePort),
		DD_DOGSTATSD_PORT.Parse(&c.StatsDPort),
	)

	return c, err
}

// Enabled returns true if DataDog is enabled
func (c *Config) Enabled() bool {
	return c.Enable
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	// If DataDog is not enabled, no need to validate further
	if !c.Enabled() {
		return nil
	}

	// Service name is required
	if len(c.Service) == 0 {
		return DD_SERVICE.ErrorEmpty()
	}

	if c.TracePort <= 0 || c.TracePort > 65535 {
		return DD_TRACE_AGENT_PORT.ErrorRange()
	}

	// StatsD port must be in the valid range
	if c.StatsDPort <= 0 || c.StatsDPort > 65535 {
		return DD_DOGSTATSD_PORT.ErrorRange()
	}

	return nil
}
