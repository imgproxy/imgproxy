package otel

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

const (
	tracesSamplerKey    = "OTEL_TRACES_SAMPLER"
	tracesSamplerArgKey = "OTEL_TRACES_SAMPLER_ARG"

	samplerTraceIDRatio = "traceidratio"
)

var (
	errNegativeTraceIDRatio       = errors.New("invalid trace ID ratio: less than 0.0")
	errGreaterThanOneTraceIDRatio = errors.New("invalid trace ID ratio: greater than 1.0")
)

type samplerArgParseError struct {
	parseErr error
}

func (e samplerArgParseError) Error() string {
	return fmt.Sprintf("parsing sampler argument: %s", e.parseErr.Error())
}

func (e samplerArgParseError) Unwrap() error {
	return e.parseErr
}

type traceIDRatioSampler struct {
	traceIDUpperBound uint64
	description       string
}

func (ts traceIDRatioSampler) ShouldSample(p sdktrace.SamplingParameters) sdktrace.SamplingResult {
	psc := trace.SpanContextFromContext(p.ParentContext)
	x := binary.BigEndian.Uint64(p.TraceID[len(p.TraceID)-8:]) >> 1
	if x < ts.traceIDUpperBound {
		return sdktrace.SamplingResult{
			Decision:   sdktrace.RecordAndSample,
			Tracestate: psc.TraceState(),
		}
	}
	return sdktrace.SamplingResult{
		Decision:   sdktrace.Drop,
		Tracestate: psc.TraceState(),
	}
}

func (ts traceIDRatioSampler) Description() string {
	return ts.description
}

func traceIDRatioBased(fraction float64) sdktrace.Sampler {
	if fraction >= 1 {
		return sdktrace.AlwaysSample()
	}

	if fraction <= 0 {
		fraction = 0
	}

	return &traceIDRatioSampler{
		traceIDUpperBound: uint64(fraction * (1 << 63)),
		description:       fmt.Sprintf("traceIDRatioBased{%g}", fraction),
	}
}

func parseTraceIDRatio(arg string) (sdktrace.Sampler, error) {
	v, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return traceIDRatioBased(1.0), samplerArgParseError{err}
	}
	if v < 0.0 {
		return traceIDRatioBased(0.0), errNegativeTraceIDRatio
	}
	if v > 1.0 {
		return traceIDRatioBased(1.0), errGreaterThanOneTraceIDRatio
	}

	return traceIDRatioBased(v), nil
}

func addTraceIDRatioSampler(opts []sdktrace.TracerProviderOption) ([]sdktrace.TracerProviderOption, error) {
	samplerName, ok := os.LookupEnv(tracesSamplerKey)
	if !ok {
		return opts, nil
	}

	if strings.ToLower(strings.TrimSpace(samplerName)) != samplerTraceIDRatio {
		return opts, nil
	}

	samplerArg, hasSamplerArg := os.LookupEnv(tracesSamplerArgKey)
	samplerArg = strings.TrimSpace(samplerArg)

	var (
		sampler sdktrace.Sampler
		err     error
	)

	if !hasSamplerArg {
		sampler = traceIDRatioBased(1.0)
	} else {
		sampler, err = parseTraceIDRatio(samplerArg)
	}

	return append(opts, sdktrace.WithSampler(sampler)), err
}
