package syslog

import (
	"context"
	"log/slog"
	"log/syslog"
	"time"
)

// Hook is a [logger.Hook] implementation for syslog.
type Hook struct {
	writer *syslog.Writer
	level  slog.Leveler
}

// NewHook creates a new syslog hook.
// It returns nil if the syslog hook is not enabled in the config.
func NewHook(config *Config) (*Hook, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	if !config.Enabled {
		return nil, nil
	}

	w, err := syslog.Dial(config.Network, config.Addr, syslog.LOG_INFO, config.Tag)
	if err != nil {
		return nil, err
	}

	return &Hook{
		writer: w,
		level:  config.Level.Level(),
	}, nil
}

func (hook *Hook) Enabled(level slog.Level) bool {
	return level >= hook.level.Level()
}

func (hook *Hook) Fire(ctx context.Context, time time.Time, lvl slog.Level, msg []byte) error {
	msgStr := string(msg)

	switch {
	case lvl < slog.LevelInfo:
		return hook.writer.Debug(msgStr)
	case lvl < slog.LevelWarn:
		return hook.writer.Info(msgStr)
	case lvl < slog.LevelError:
		return hook.writer.Warning(msgStr)
	default:
		return hook.writer.Err(msgStr)
	}
}
