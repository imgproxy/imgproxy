package logger

import (
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

var handlerBenchmarkMsg = "test message"
var handlerBenchmarkAttrs = []any{
	slog.String("string", "value"),
	slog.Int("int", -100),
	slog.Uint64("uint64", 200),
	slog.Float64("float64", 3.14),
	slog.Bool("bool", true),
	slog.Time("time", time.Now()),
	slog.Duration("duration", time.Minute+time.Second),
	slog.Group("group", "group_key", "group_value"),
	slog.Any("err", errors.New("error value")),
	slog.Any("any", struct {
		Field1 string
		Field2 int
	}{"value", 42}),
}

func BenchmarkFormatterPretty(b *testing.B) {
	testHandler := NewHandler(io.Discard, &Config{
		Level:  slog.LevelDebug,
		Format: FormatPretty,
	})
	testLogger := slog.New(testHandler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testLogger.Info(
			handlerBenchmarkMsg,
			handlerBenchmarkAttrs...,
		)
	}
}

func BenchmarkFormatterStructured(b *testing.B) {
	testHandler := NewHandler(io.Discard, &Config{
		Level:  slog.LevelDebug,
		Format: FormatStructured,
	})
	testLogger := slog.New(testHandler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testLogger.Info(
			handlerBenchmarkMsg,
			handlerBenchmarkAttrs...,
		)
	}
}

func BenchmarkFormatterJSON(b *testing.B) {
	testHandler := NewHandler(io.Discard, &Config{
		Level:  slog.LevelDebug,
		Format: FormatJSON,
	})
	testLogger := slog.New(testHandler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testLogger.Info(
			handlerBenchmarkMsg,
			handlerBenchmarkAttrs...,
		)
	}
}

func BenchmarkNativeText(b *testing.B) {
	testHandler := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})
	testLogger := slog.New(testHandler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testLogger.Info(
			handlerBenchmarkMsg,
			handlerBenchmarkAttrs...,
		)
	}
}

func BenchmarkNativeJSON(b *testing.B) {
	testHandler := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})
	testLogger := slog.New(testHandler)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testLogger.Info(
			handlerBenchmarkMsg,
			handlerBenchmarkAttrs...,
		)
	}
}
