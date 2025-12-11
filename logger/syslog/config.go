package syslog

import (
	"errors"
	"log/slog"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	syslogLevelMap = map[string]slog.Level{
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
		"crit":  slog.LevelError + 8,
	}

	IMGPROXY_SYSLOG_ENABLE  = env.Bool("IMGPROXY_SYSLOG_ENABLE")
	IMGPROXY_SYSLOG_LEVEL   = env.Enum("IMGPROXY_SYSLOG_LEVEL", syslogLevelMap)
	IMGPROXY_SYSLOG_NETWORK = env.String("IMGPROXY_SYSLOG_NETWORK")
	IMGPROXY_SYSLOG_ADDRESS = env.String("IMGPROXY_SYSLOG_ADDRESS")
	IMGPROXY_SYSLOG_TAG     = env.String("IMGPROXY_SYSLOG_TAG")
)

type Config struct {
	Enabled bool
	Level   slog.Leveler
	Network string
	Addr    string
	Tag     string
}

func NewDefaultConfig() Config {
	return Config{
		Enabled: false,
		Level:   slog.LevelInfo,
		Tag:     "imgproxy",
	}
}

func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	level := c.Level.Level()

	err := errors.Join(
		IMGPROXY_SYSLOG_ENABLE.Parse(&c.Enabled),
		IMGPROXY_SYSLOG_NETWORK.Parse(&c.Network),
		IMGPROXY_SYSLOG_ADDRESS.Parse(&c.Addr),
		IMGPROXY_SYSLOG_TAG.Parse(&c.Tag),
		IMGPROXY_SYSLOG_LEVEL.Parse(&level),
	)

	c.Level = level

	return c, err
}

func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Network != "" && c.Addr == "" {
		return errors.New("syslog address is required if syslog network is set")
	}

	return nil
}
