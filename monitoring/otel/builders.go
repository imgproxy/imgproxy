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

// buildProtocolExporter builds trace and metric exporters based on the provided configuration.
func buildProtocolExporter(config *Config) (te *otlptrace.Exporter, me sdkmetric.Exporter, err error) {
	switch config.Protocol {
	case "grpc":
		te, me, err = buildGRPCExporters(config)
	case "http/protobuf", "http", "https":
		te, me, err = buildHTTPExporters(config)
	default:
		err = fmt.Errorf("unsupported OpenTelemetry protocol: %s", config.Protocol)
	}

	return
}

// buildGRPCExporters builds GRPC exporters based on the provided configuration.
func buildGRPCExporters(config *Config) (*otlptrace.Exporter, sdkmetric.Exporter, error) {
	tracerOpts := []otlptracegrpc.Option{}
	meterOpts := []otlpmetricgrpc.Option{}

	if tlsConf, err := buildTLSConfig(config); tlsConf != nil && err == nil {
		creds := credentials.NewTLS(tlsConf)
		tracerOpts = append(tracerOpts, otlptracegrpc.WithTLSCredentials(creds))
		meterOpts = append(meterOpts, otlpmetricgrpc.WithTLSCredentials(creds))
	} else if err != nil {
		return nil, nil, err
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	trctx, trcancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.TracesConnTimeout),
	)
	defer trcancel()

	traceExporter, err := otlptracegrpc.New(trctx, tracerOpts...)
	if err != nil {
		err = fmt.Errorf("can't connect to OpenTelemetry collector: %w", err)
	}

	if !config.EnableMetrics {
		return traceExporter, nil, err
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	mtctx, mtcancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.MetricsConnTimeout),
	)
	defer mtcancel()

	metricExporter, err := otlpmetricgrpc.New(mtctx, meterOpts...)
	if err != nil {
		err = fmt.Errorf("can't connect to OpenTelemetry collector: %w", err)
	}

	return traceExporter, metricExporter, err
}

// buildHTTPExporters builds HTTP exporters based on the provided configuration.
func buildHTTPExporters(config *Config) (*otlptrace.Exporter, sdkmetric.Exporter, error) {
	tracerOpts := []otlptracehttp.Option{}
	meterOpts := []otlpmetrichttp.Option{}

	if tlsConf, err := buildTLSConfig(config); tlsConf != nil && err == nil {
		tracerOpts = append(tracerOpts, otlptracehttp.WithTLSClientConfig(tlsConf))
		meterOpts = append(meterOpts, otlpmetrichttp.WithTLSClientConfig(tlsConf))
	} else if err != nil {
		return nil, nil, err
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	trctx, trcancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.TracesConnTimeout),
	)
	defer trcancel()

	traceExporter, err := otlptracehttp.New(trctx, tracerOpts...)
	if err != nil {
		err = fmt.Errorf("can't connect to OpenTelemetry collector: %w", err)
	}

	if !config.EnableMetrics {
		return traceExporter, nil, err
	}

	// This context limits connect timeout, not the whole lifetime of the exporter
	mtctx, mtcancel := context.WithTimeout(
		context.Background(), withDefaultTimeout(config, config.MetricsConnTimeout),
	)
	defer mtcancel()

	metricExporter, err := otlpmetrichttp.New(mtctx, meterOpts...)
	if err != nil {
		err = fmt.Errorf("can't connect to OpenTelemetry collector: %w", err)
	}

	return traceExporter, metricExporter, err
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
func buildGRPCLogExporter(config *Config) (sdklog.Exporter, error) {
	opts := []otlploggrpc.Option{}

	if tlsConf, err := buildTLSConfig(config); tlsConf != nil && err == nil {
		creds := credentials.NewTLS(tlsConf)
		opts = append(opts, otlploggrpc.WithTLSCredentials(creds))
	} else if err != nil {
		return nil, err
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
func buildHTTPLogExporter(config *Config) (sdklog.Exporter, error) {
	opts := []otlploghttp.Option{}

	if tlsConf, err := buildTLSConfig(config); tlsConf != nil && err == nil {
		opts = append(opts, otlploghttp.WithTLSClientConfig(tlsConf))
	} else if err != nil {
		return nil, err
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

// buildLogExporter builds a log exporter based on the protocol in the configuration.
func buildLogExporter(config *Config) (sdklog.Exporter, error) {
	if !config.EnableLogs {
		return nil, nil
	}

	switch config.Protocol {
	case "grpc":
		return buildGRPCLogExporter(config)
	case "http/protobuf", "http", "https":
		return buildHTTPLogExporter(config)
	default:
		return nil, fmt.Errorf("unsupported OpenTelemetry protocol for logs: %s", config.Protocol)
	}
}
