package logger

import (
	"errors"
	"log/slog"
	"os"

	"github.com/mattn/go-isatty"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/logger/syslog"
)

var (
	logFormatMap = map[string]Format{
		"pretty":     FormatPretty,
		"structured": FormatStructured,
		"json":       FormatJSON,
		"gcp":        FormatGCP,
	}

	logLevelMap = map[string]slog.Leveler{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	IMGPROXY_LOG_FORMAT = env.Enum("IMGPROXY_LOG_FORMAT", logFormatMap)
	IMGPROXY_LOG_LEVEL  = env.Enum("IMGPROXY_LOG_LEVEL", logLevelMap)
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

	_, slErr := syslog.LoadConfigFromEnv(&o.Syslog)

	err := errors.Join(
		slErr,
		IMGPROXY_LOG_FORMAT.Parse(&o.Format),
		IMGPROXY_LOG_LEVEL.Parse(&o.Level),
	)

	return o, err
}

func (c *Config) Validate() error {
	return c.Syslog.Validate()
}
