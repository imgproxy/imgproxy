package logger

import (
	"errors"
	"log/slog"
	"os"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/logger/syslog"
)

var (
	IMGPROXY_LOG_FORMAT = env.Describe("IMGPROXY_LOG_FORMAT", "pretty|structured|json|gcp")
	IMGPROXY_LOG_LEVEL  = env.Describe("IMGPROXY_LOG_LEVEL", "debug|info|warn|error")
)

type Config struct {
	Level  slog.Leveler
	Format Format

	Syslog syslog.Config
}

func NewDefaultConfig() Config {
	o := Config{
		Level:  slog.LevelInfo,
		Format: FormatStructured,

		Syslog: syslog.NewDefaultConfig(),
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		o.Format = FormatPretty
	}

	return o
}

func LoadConfigFromEnv(o *Config) (*Config, error) {
	o = ensure.Ensure(o, NewDefaultConfig)

	var logFormat, logLevel string

	_, slErr := syslog.LoadConfigFromEnv(&o.Syslog)

	err := errors.Join(
		slErr,
		env.String(&logFormat, IMGPROXY_LOG_FORMAT),
		env.String(&logLevel, IMGPROXY_LOG_LEVEL),
	)

	if logFormat != "" {
		o.Format = parseFormat(logFormat)
	}

	if logLevel != "" {
		o.Level = parseLevel(logLevel)
	}

	// Load syslog config

	return o, err
}

func (c *Config) Validate() error {
	return c.Syslog.Validate()
}

func parseFormat(str string) Format {
	switch str {
	case "pretty":
		return FormatPretty
	case "structured":
		return FormatStructured
	case "json":
		return FormatJSON
	case "gcp":
		return FormatGCP
	default:
		if isatty.IsTerminal(os.Stdout.Fd()) {
			return FormatPretty
		}
		return FormatStructured
	}
}

func parseLevel(str string) slog.Level {
	switch strings.ToLower(str) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
