package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/mattn/go-isatty"

	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/logger/syslog"
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

func LoadConfigFromEnv(o *Config) *Config {
	o = ensure.Ensure(o, NewDefaultConfig)

	var logFormat, logLevel string
	configurators.String(&logFormat, "IMGPROXY_LOG_FORMAT")
	configurators.String(&logLevel, "IMGPROXY_LOG_LEVEL")

	if logFormat != "" {
		o.Format = parseFormat(logFormat)
	}
	if logLevel != "" {
		o.Level = parseLevel(logLevel)
	}

	// Load syslog config
	syslog.LoadConfigFromEnv(&o.Syslog)

	return o
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
