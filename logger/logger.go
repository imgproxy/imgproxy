package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config/configurators"
	"github.com/imgproxy/imgproxy/v3/logger/gliblog"
	"github.com/imgproxy/imgproxy/v3/logger/syslog"
)

// We store a [Handler] instance here so that we can restore it in [Unmute]
var handler *Handler

// init initializes the default logger
func init() {
	cfg := NewDefaultConfig()
	handler = NewHandler(os.Stdout, &cfg)
	setDefaultHandler(handler)
}

// Init creates a logger and sets it as the default log/slog logger
func Init(config *Config) error {
	if err := config.Validate(); err != nil {
		return err
	}

	handler = NewHandler(os.Stdout, config)
	setDefaultHandler(handler)

	gliblog.Init()

	if config.Syslog.Enabled {
		slHook, err := syslog.NewHook(&config.Syslog)
		if err != nil {
			return fmt.Errorf("unable to connect to syslog daemon: %w", err)
		}
		if slHook != nil {
			AddHook(slHook)
		}
	}

	return nil
}

func AddHook(hook Hook) {
	if handler != nil {
		handler.AddHook(hook)
	}
}

func Fatal(msg string, args ...any) {
	slog.Log(context.Background(), LevelCritical, msg, args...)
	os.Exit(1)
}

// Deprecated prints a deprecation warning message.
// If the IMGPROXY_FAIL_ON_DEPRECATION environment variable is truthy,
// it prints an error message and exits the program.
func Deprecated(deprecation, replacement string, additional ...string) {
	msg := fmt.Sprintf("%s is deprecated, use %s instead", deprecation, replacement)

	if len(additional) > 0 {
		msg += ". " + strings.Join(additional, ". ")
	}

	shouldFail := false
	configurators.Bool(&shouldFail, "IMGPROXY_FAIL_ON_DEPRECATION")

	if shouldFail {
		Fatal(msg)
	} else {
		slog.Warn(msg)
	}
}

// Mute sets the default logger to a discard logger muting all log output
func Mute() {
	setDefaultHandler(slog.DiscardHandler)
}

// Unmute restores the default logger to the one created in [Init]
func Unmute() {
	setDefaultHandler(handler)
}

func setDefaultHandler(h slog.Handler) {
	slog.SetDefault(slog.New(h))
}
