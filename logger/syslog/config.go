package syslog

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/ensure"
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

func LoadConfigFromEnv(c *Config) *Config {
	c = ensure.Ensure(c, NewDefaultConfig)

	configurators.Bool(&c.Enabled, "IMGPROXY_SYSLOG_ENABLE")

	configurators.String(&c.Network, "IMGPROXY_SYSLOG_NETWORK")
	configurators.String(&c.Addr, "IMGPROXY_SYSLOG_ADDRESS")
	configurators.String(&c.Tag, "IMGPROXY_SYSLOG_TAG")

	var levelStr string
	configurators.String(&levelStr, "IMGPROXY_SYSLOG_LEVEL")

	if levelStr != "" {
		c.Level = parseLevel(levelStr)
	}

	return c
}

func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	if c.Network != "" && c.Addr == "" {
		return errors.New("Syslog address is required if syslog network is set")
	}

	return nil
}

func parseLevel(str string) slog.Level {
	switch strings.ToLower(str) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "crit":
		return slog.LevelError + 8
	default:
		slog.Warn(fmt.Sprintf("Syslog level '%s' is invalid, 'info' is used", str))
		return slog.LevelInfo
	}
}
