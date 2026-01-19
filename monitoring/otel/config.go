package otel

import (
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

const (
	otelDocsUrl = "https://opentelemetry.io/docs/languages/sdk-configuration/otlp-exporter/"
)

var (
	logLevelMap = map[string]slog.Leveler{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	protocolMap = map[string]string{
		"grpc":          "grpc",
		"http/protobuf": "http/protobuf",
		"http":          "http/protobuf",
		"https":         "http/protobuf",
	}

	IMGPROXY_OPEN_TELEMETRY_ENABLE             = env.Bool("IMGPROXY_OPEN_TELEMETRY_ENABLE")
	IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS     = env.Bool("IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS")
	IMGPROXY_OPEN_TELEMETRY_ENABLE_LOGS        = env.Bool("IMGPROXY_OPEN_TELEMETRY_ENABLE_LOGS")
	IMGPROXY_OPEN_TELEMETRY_LOGGER_NAME        = env.String("IMGPROXY_OPEN_TELEMETRY_LOGGER_NAME")
	IMGPROXY_OPEN_TELEMETRY_SERVER_CERT        = env.String("IMGPROXY_OPEN_TELEMETRY_SERVER_CERT")
	IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT        = env.String("IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT")
	IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY         = env.String("IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY")
	IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR = env.String("IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR")
	IMGPROXY_OPEN_TELEMETRY_PROPAGATE_EXTERNAL = env.Bool("IMGPROXY_OPEN_TELEMETRY_PROPAGATE_EXTERNAL")

	// OTEL_EXPORTER_OTLP_PROTOCOL Those are OpenTelemetry SDK environment variables
	OTEL_EXPORTER_OTLP_PROTOCOL         = env.Enum("OTEL_EXPORTER_OTLP_PROTOCOL", protocolMap).WithDocsURL(otelDocsUrl)
	OTEL_EXPORTER_OTLP_TRACES_PROTOCOL  = env.Enum("OTEL_EXPORTER_OTLP_TRACES_PROTOCOL", protocolMap).WithDocsURL(otelDocsUrl)  //nolint:lll
	OTEL_EXPORTER_OTLP_METRICS_PROTOCOL = env.Enum("OTEL_EXPORTER_OTLP_METRICS_PROTOCOL", protocolMap).WithDocsURL(otelDocsUrl) //nolint:lll
	OTEL_EXPORTER_OTLP_LOGS_PROTOCOL    = env.Enum("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL", protocolMap).WithDocsURL(otelDocsUrl)    //nolint:lll
	OTEL_EXPORTER_OTLP_TIMEOUT          = env.DurationMillis("OTEL_EXPORTER_OTLP_TIMEOUT").WithDocsURL(otelDocsUrl)
	OTEL_EXPORTER_OTLP_TRACES_TIMEOUT   = env.DurationMillis("OTEL_EXPORTER_OTLP_TRACES_TIMEOUT").WithDocsURL(otelDocsUrl)
	OTEL_EXPORTER_OTLP_METRICS_TIMEOUT  = env.DurationMillis("OTEL_EXPORTER_OTLP_METRICS_TIMEOUT").WithDocsURL(otelDocsUrl)
	OTEL_EXPORTER_OTLP_LOGS_TIMEOUT     = env.DurationMillis("OTEL_EXPORTER_OTLP_LOGS_TIMEOUT").WithDocsURL(otelDocsUrl)
	OTEL_PROPAGATORS                    = env.StringSlice("OTEL_PROPAGATORS").WithDocsURL(otelDocsUrl)
	OTEL_SERVICE_NAME                   = env.String("OTEL_SERVICE_NAME").WithDocsURL(otelDocsUrl)
	OTEL_LOG_LEVEL                      = env.Enum("OTEL_LOG_LEVEL", logLevelMap).WithDocsURL(otelDocsUrl)
)

// Config holds the configuration for OpenTelemetry monitoring
type Config struct {
	Enable           bool         // Enable OpenTelemetry tracing and metrics
	EnableMetrics    bool         // Enable OpenTelemetry metrics collection
	EnableLogs       bool         // Enable OpenTelemetry log export
	LogLevel         slog.Leveler // Minimum log level for OTLP export (defaults to logger level if not set)
	LoggerName       string       // Instrumentation scope name for logs
	ServerCert       []byte       // Server certificate for TLS connection
	ClientCert       []byte       // Client certificate for TLS connection
	ClientKey        []byte       // Client key for TLS connection
	TraceIDGenerator string       // Trace ID generator type (e.g., "xray", "random")
	PropagateExt     bool         // Enable propagation of tracing headers for external services

	Protocol           string        // Protocol to use for OTLP exporter (grpc, http/protobuf, http, https)
	TracesProtocol     string        // Protocol to use for traces OTLP exporter
	MetricsProtocol    string        // Protocol to use for metrics OTLP exporter
	LogsProtocol       string        // Protocol to use for logs OTLP exporter
	ConnTimeout        time.Duration // Connection timeout for OTLP exporter
	MetricsConnTimeout time.Duration // Connection timeout for metrics exporter
	TracesConnTimeout  time.Duration // Connection timeout for traces exporter
	LogsConnTimeout    time.Duration // Connection timeout for logs exporter
	Propagators        []string      // List of propagators to use

	MetricsInterval time.Duration // Interval for sending metrics to OpenTelemetry collector
}

// NewDefaultConfig returns a new default configuration for OpenTelemetry monitoring
func NewDefaultConfig() Config {
	return Config{
		Enable:             false,
		EnableMetrics:      false,
		EnableLogs:         false,
		LoggerName:         "imgproxy",
		ServerCert:         nil,
		ClientCert:         nil,
		ClientKey:          nil,
		TraceIDGenerator:   "xray",
		PropagateExt:       false,
		Protocol:           "grpc",
		TracesProtocol:     "",
		MetricsProtocol:    "",
		LogsProtocol:       "",
		ConnTimeout:        10_000 * time.Millisecond,
		MetricsConnTimeout: 0,
		TracesConnTimeout:  0,
		LogsConnTimeout:    0,
		Propagators:        []string{"tracecontext", "baggage"},
		MetricsInterval:    10 * time.Second,
	}
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	var serverCert, clientCert, clientKey string

	err := errors.Join(
		IMGPROXY_OPEN_TELEMETRY_ENABLE.Parse(&c.Enable),
		IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS.Parse(&c.EnableMetrics),
		IMGPROXY_OPEN_TELEMETRY_ENABLE_LOGS.Parse(&c.EnableLogs),
		IMGPROXY_OPEN_TELEMETRY_LOGGER_NAME.Parse(&c.LoggerName),
		IMGPROXY_OPEN_TELEMETRY_SERVER_CERT.Parse(&serverCert),
		IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT.Parse(&clientCert),
		IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY.Parse(&clientKey),
		IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR.Parse(&c.TraceIDGenerator),
		IMGPROXY_OPEN_TELEMETRY_PROPAGATE_EXTERNAL.Parse(&c.PropagateExt),
		OTEL_EXPORTER_OTLP_PROTOCOL.Parse(&c.Protocol),
		OTEL_EXPORTER_OTLP_TRACES_PROTOCOL.Parse(&c.TracesProtocol),
		OTEL_EXPORTER_OTLP_METRICS_PROTOCOL.Parse(&c.MetricsProtocol),
		OTEL_EXPORTER_OTLP_LOGS_PROTOCOL.Parse(&c.LogsProtocol),
		OTEL_EXPORTER_OTLP_TIMEOUT.Parse(&c.ConnTimeout),
		OTEL_EXPORTER_OTLP_TRACES_TIMEOUT.Parse(&c.TracesConnTimeout),
		OTEL_EXPORTER_OTLP_METRICS_TIMEOUT.Parse(&c.MetricsConnTimeout),
		OTEL_EXPORTER_OTLP_LOGS_TIMEOUT.Parse(&c.LogsConnTimeout),
		OTEL_PROPAGATORS.Parse(&c.Propagators),
		OTEL_LOG_LEVEL.Parse(&c.LogLevel),
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

	// Protocol must be set
	if c.Protocol == "" {
		return OTEL_EXPORTER_OTLP_PROTOCOL.ErrorEmpty()
	}

	return nil
}

func prepareKeyCert(str string) []byte {
	return []byte(strings.ReplaceAll(str, `\n`, "\n"))
}
