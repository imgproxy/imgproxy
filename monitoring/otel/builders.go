package otel

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"google.golang.org/grpc/credentials"
)

const (
	grpcProtocol     = "grpc"
	protobufProtocol = "http/protobuf"
)

// buildExporters builds trace, metric, and log exporters based on the provided configuration.
func buildExporters(config *Config) (
	te *otlptrace.Exporter,
	me sdkmetric.Exporter,
	le sdklog.Exporter,
	err error,
) {
	tlsConf, err := buildTLSConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	te, err = buildTraceExporter(config, tlsConf)
	if err != nil {
		return nil, nil, nil, err
	}

	me, err = buildMetricExporter(config, tlsConf)
	if err != nil {
		return nil, nil, nil, err
	}

	le, err = buildLogExporter(config, tlsConf)
	return te, me, le, err
}

// getProtocol returns the specific protocol if set, otherwise falls back to the general protocol.
func getProtocol(specific, fallback string) string {
	if specific == "" {
		return fallback
	}
	return specific
}

// buildTraceExporter builds a trace exporter based on the protocol in the configuration.
func buildTraceExporter(config *Config, tlsConf *tls.Config) (*otlptrace.Exporter, error) {
	protocol := getProtocol(config.TracesProtocol, config.Protocol)

	switch protocol {
	case grpcProtocol:
		return buildGRPCTraceExporter(config, tlsConf)
	case protobufProtocol:
		return buildHTTPTraceExporter(config, tlsConf)
	default:
		return nil, fmt.Errorf("unsupported OpenTelemetry protocol for traces: %s", protocol)
	}
}

// buildMetricExporter builds a metric exporter based on the protocol in the configuration.
func buildMetricExporter(config *Config, tlsConf *tls.Config) (sdkmetric.Exporter, error) {
	if !config.EnableMetrics {
		return nil, nil
	}

	protocol := getProtocol(config.MetricsProtocol, config.Protocol)

	switch protocol {
	case grpcProtocol:
		return buildGRPCMetricExporter(config, tlsConf)
	case protobufProtocol:
		return buildHTTPMetricExporter(config, tlsConf)
	default:
		return nil, fmt.Errorf("unsupported OpenTelemetry protocol for metrics: %s", protocol)
	}
}

// buildLogExporter builds a log exporter based on the protocol in the configuration.
func buildLogExporter(config *Config, tlsConf *tls.Config) (sdklog.Exporter, error) {
	if !config.EnableLogs {
		return nil, nil
	}

	protocol := getProtocol(config.LogsProtocol, config.Protocol)

	switch protocol {
	case grpcProtocol:
		return buildGRPCLogExporter(config, tlsConf)
	case protobufProtocol:
		return buildHTTPLogExporter(config, tlsConf)
	default:
		return nil, fmt.Errorf("unsupported OpenTelemetry protocol for logs: %s", protocol)
	}
}

// buildGRPCTraceExporter builds a GRPC trace exporter based on the provided configuration.
func buildGRPCTraceExporter(config *Config, tlsConf *tls.Config) (*otlptrace.Exporter, error) {
	opts := []otlptracegrpc.Option{}

	if tlsConf != nil {
		creds := credentials.NewTLS(tlsConf)
		opts = append(opts, otlptracegrpc.WithTLSCredentials(creds))
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	ctx, cancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.TracesConnTimeout),
	)
	defer cancel()

	traceExporter, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("can't connect to OpenTelemetry collector for traces: %w", err)
	}

	return traceExporter, nil
}

// buildGRPCMetricExporter builds a GRPC metric exporter based on the provided configuration.
func buildGRPCMetricExporter(config *Config, tlsConf *tls.Config) (sdkmetric.Exporter, error) {
	opts := []otlpmetricgrpc.Option{}

	if tlsConf != nil {
		creds := credentials.NewTLS(tlsConf)
		opts = append(opts, otlpmetricgrpc.WithTLSCredentials(creds))
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	ctx, cancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.MetricsConnTimeout),
	)
	defer cancel()

	metricExporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("can't connect to OpenTelemetry collector for metrics: %w", err)
	}

	return metricExporter, nil
}

// buildHTTPTraceExporter builds an HTTP trace exporter based on the provided configuration.
func buildHTTPTraceExporter(config *Config, tlsConf *tls.Config) (*otlptrace.Exporter, error) {
	opts := []otlptracehttp.Option{}

	if tlsConf != nil {
		opts = append(opts, otlptracehttp.WithTLSClientConfig(tlsConf))
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	ctx, cancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.TracesConnTimeout),
	)
	defer cancel()

	traceExporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("can't connect to OpenTelemetry collector for traces: %w", err)
	}

	return traceExporter, nil
}

// buildHTTPMetricExporter builds an HTTP metric exporter based on the provided configuration.
func buildHTTPMetricExporter(config *Config, tlsConf *tls.Config) (sdkmetric.Exporter, error) {
	opts := []otlpmetrichttp.Option{}

	if tlsConf != nil {
		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(tlsConf))
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	ctx, cancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.MetricsConnTimeout),
	)
	defer cancel()

	metricExporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("can't connect to OpenTelemetry collector for metrics: %w", err)
	}

	return metricExporter, nil
}

// buildTLSConfig constructs a tls.Config based on the provided configuration.
func buildTLSConfig(config *Config) (*tls.Config, error) {
	// If no server certificate is provided, we assume no TLS is needed
	if len(config.ServerCert) == 0 {
		return nil, nil
	}

	// Attach root CAs
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(config.ServerCert) {
		return nil, errors.New("can't load OpenTelemetry server cert")
	}

	// Package default is 1.2
	//nolint:gosec
	tlsConf := tls.Config{RootCAs: certPool}

	// If there is not client cert or key, return the config with only root CAs
	if len(config.ClientCert) == 0 || len(config.ClientKey) == 0 {
		return &tlsConf, nil
	}

	cert, err := tls.X509KeyPair(config.ClientCert, config.ClientKey)
	if err != nil {
		return nil, fmt.Errorf("can't load OpenTelemetry client cert/key pair: %w", err)
	}

	tlsConf.Certificates = []tls.Certificate{cert}

	return &tlsConf, nil
}

func withDefaultTimeout(config *Config, timeout time.Duration) time.Duration {
	// In case, timeout is zero or negative we assume it was not set
	// (or was set to invalid value) and use the default timeout
	if timeout <= 0 {
		return config.ConnTimeout
	}
	return timeout
}

// buildGRPCLogExporter builds a GRPC log exporter based on the provided configuration.
func buildGRPCLogExporter(config *Config, tlsConf *tls.Config) (sdklog.Exporter, error) {
	opts := []otlploggrpc.Option{}

	if tlsConf != nil {
		creds := credentials.NewTLS(tlsConf)
		opts = append(opts, otlploggrpc.WithTLSCredentials(creds))
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	ctx, cancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.LogsConnTimeout),
	)
	defer cancel()

	logExporter, err := otlploggrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("can't connect to OpenTelemetry collector for logs: %w", err)
	}

	return logExporter, nil
}

// buildHTTPLogExporter builds an HTTP log exporter based on the provided configuration.
func buildHTTPLogExporter(config *Config, tlsConf *tls.Config) (sdklog.Exporter, error) {
	opts := []otlploghttp.Option{}

	if tlsConf != nil {
		opts = append(opts, otlploghttp.WithTLSClientConfig(tlsConf))
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	ctx, cancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.LogsConnTimeout),
	)
	defer cancel()

	logExporter, err := otlploghttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("can't connect to OpenTelemetry collector for logs: %w", err)
	}

	return logExporter, nil
}
