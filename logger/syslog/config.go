package syslog

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_SYSLOG_ENABLE  = env.Describe("IMGPROXY_SYSLOG_ENABLE", "boolean")
	IMGPROXY_SYSLOG_LEVEL   = env.Describe("IMGPROXY_SYSLOG_LEVEL", "debug|info|warn|error|crit")
	IMGPROXY_SYSLOG_NETWORK = env.Describe("IMGPROXY_SYSLOG_NETWORK", "string")
	IMGPROXY_SYSLOG_ADDRESS = env.Describe("IMGPROXY_SYSLOG_ADDRESS", "string")
	IMGPROXY_SYSLOG_TAG     = env.Describe("IMGPROXY_SYSLOG_TAG", "string")
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

	var levelStr string

	err := errors.Join(
		env.Bool(&c.Enabled, IMGPROXY_SYSLOG_ENABLE),
		env.String(&c.Network, IMGPROXY_SYSLOG_NETWORK),
		env.String(&c.Addr, IMGPROXY_SYSLOG_ADDRESS),
		env.String(&c.Tag, IMGPROXY_SYSLOG_TAG),
		env.String(&levelStr, IMGPROXY_SYSLOG_LEVEL),
	)

	if levelStr != "" {
		c.Level = parseLevel(levelStr)
	}

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
