package otel

import (
	"errors"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_OPEN_TELEMETRY_ENABLE             = env.Describe("IMGPROXY_OPEN_TELEMETRY_ENABLE", "boolean")
	IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS     = env.Describe("IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS", "boolean")
	IMGPROXY_OPEN_TELEMETRY_SERVER_CERT        = env.Describe("IMGPROXY_OPEN_TELEMETRY_SERVER_CERT", "string")
	IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT        = env.Describe("IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT", "string")
	IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY         = env.Describe("IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY", "string")
	IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR = env.Describe("IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR", "xray|random")
	IMGPROXY_OPEN_TELEMETRY_PROPAGATE_EXTERNAL = env.Describe("IMGPROXY_OPEN_TELEMETRY_PROPAGATE_EXTERNAL", "boolean")

	// Those are OpenTelemetry SDK environment variables
	OTEL_EXPORTER_OTLP_PROTOCOL        = env.Describe("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc|http/protobuf|http|https")
	OTEL_EXPORTER_OTLP_TIMEOUT         = env.Describe("OTEL_EXPORTER_OTLP_TIMEOUT", "milliseconds")
	OTEL_EXPORTER_OTLP_TRACES_TIMEOUT  = env.Describe("OTEL_EXPORTER_OTLP_TRACES_TIMEOUT", "milliseconds")
	OTEL_EXPORTER_OTLP_METRICS_TIMEOUT = env.Describe("OTEL_EXPORTER_OTLP_METRICS_TIMEOUT", "milliseconds")
	OTEL_PROPAGATORS                   = env.Describe("OTEL_PROPAGATORS", "comma-separated list of propagators")
	OTEL_SERVICE_NAME                  = env.Describe("OTEL_SERVICE_NAME", "string") // This is used during initialization
)

// Config holds the configuration for OpenTelemetry monitoring
type Config struct {
	Enable           bool   // Enable OpenTelemetry tracing and metrics
	EnableMetrics    bool   // Enable OpenTelemetry metrics collection
	ServerCert       []byte // Server certificate for TLS connection
	ClientCert       []byte // Client certificate for TLS connection
	ClientKey        []byte // Client key for TLS connection
	TraceIDGenerator string // Trace ID generator type (e.g., "xray", "random")
	PropagateExt     bool   // Enable propagation of tracing headers for external services

	Protocol           string        // Protocol to use for OTLP exporter (grpc, http/protobuf, http, https)
	ConnTimeout        time.Duration // Connection timeout for OTLP exporter
	MetricsConnTimeout time.Duration // Connection timeout for metrics exporter
	TracesConnTimeout  time.Duration // Connection timeout for traces exporter
	Propagators        []string      // List of propagators to use

	MetricsInterval time.Duration // Interval for sending metrics to OpenTelemetry collector
}

// NewDefaultConfig returns a new default configuration for OpenTelemetry monitoring
func NewDefaultConfig() Config {
	return Config{
		Enable:             false,
		EnableMetrics:      false,
		ServerCert:         nil,
		ClientCert:         nil,
		ClientKey:          nil,
		TraceIDGenerator:   "xray",
		PropagateExt:       false,
		Protocol:           "grpc",
		ConnTimeout:        10_000 * time.Millisecond,
		MetricsConnTimeout: 0,
		TracesConnTimeout:  0,
		Propagators:        []string{},
		MetricsInterval:    10 * time.Second,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	var serverCert, clientCert, clientKey string

	err := errors.Join(
		env.Bool(&c.Enable, IMGPROXY_OPEN_TELEMETRY_ENABLE),
		env.Bool(&c.EnableMetrics, IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS),
		env.String(&serverCert, IMGPROXY_OPEN_TELEMETRY_SERVER_CERT),
		env.String(&clientCert, IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT),
		env.String(&clientKey, IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY),
		env.String(&c.TraceIDGenerator, IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR),
		env.Bool(&c.PropagateExt, IMGPROXY_OPEN_TELEMETRY_PROPAGATE_EXTERNAL),
		env.String(&c.Protocol, OTEL_EXPORTER_OTLP_PROTOCOL),
		env.DurationMils(&c.ConnTimeout, OTEL_EXPORTER_OTLP_TIMEOUT),
		env.DurationMils(&c.TracesConnTimeout, OTEL_EXPORTER_OTLP_TRACES_TIMEOUT),
		env.DurationMils(&c.MetricsConnTimeout, OTEL_EXPORTER_OTLP_METRICS_TIMEOUT),
		env.StringSlice(&c.Propagators, OTEL_PROPAGATORS),
	)

	c.ServerCert = prepareKeyCert(serverCert)
	c.ClientCert = prepareKeyCert(clientCert)
	c.ClientKey = prepareKeyCert(clientKey)

	return c, err
}

func (c *Config) Enabled() bool {
	return c.Enable
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	if !c.Enabled() {
		return nil
	}

	// Timeout should be valid
	if c.ConnTimeout <= 0 {
		return OTEL_EXPORTER_OTLP_TIMEOUT.ErrorZeroOrNegative()
	}

	return nil
}

func prepareKeyCert(str string) []byte {
	return []byte(strings.ReplaceAll(str, `\n`, "\n"))
}
