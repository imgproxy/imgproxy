package datadog

import (
	"errors"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_DATADOG_ENABLE                    = env.Describe("IMGPROXY_DATADOG_ENABLE", "boolean")
	IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS = env.Describe("IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS", "boolean")

	DD_SERVICE            = env.Describe("DD_SERVICE", "string")
	DD_TRACE_STARTUP_LOGS = env.Describe("DD_TRACE_STARTUP_LOGS", "boolean")
	DD_AGENT_HOST         = env.Describe("DD_AGENT_HOST", "host")
	DD_TRACE_AGENT_PORT   = env.Describe("DD_TRACE_AGENT_PORT", "port")
	DD_DOGSTATSD_PORT     = env.Describe("DD_DOGSTATSD_PORT", "port")
)

// Config holds the configuration for DataDog monitoring
type Config struct {
	Enable           bool          // Enable DataDog tracing
	EnableMetrics    bool          // Enable DataDog metrics collection
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
		env.Bool(&c.Enable, IMGPROXY_DATADOG_ENABLE),
		env.Bool(&c.EnableMetrics, IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS),
		env.String(&c.Service, DD_SERVICE),
		env.Bool(&c.TraceStartupLogs, DD_TRACE_STARTUP_LOGS),
		env.String(&c.AgentHost, DD_AGENT_HOST),
		env.Int(&c.TracePort, DD_TRACE_AGENT_PORT),
		env.Int(&c.StatsDPort, DD_DOGSTATSD_PORT),
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
