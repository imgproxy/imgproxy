package otel

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	imgproxylogger "github.com/imgproxy/imgproxy/v3/logger"
)

// logHook is a logger.Hook that exports logs to OpenTelemetry.
type logHook struct {
	provider *sdklog.LoggerProvider
	logger   log.Logger
	level    slog.Level
}

// newLogHook creates a new OpenTelemetry log hook.
func newLogHook(provider *sdklog.LoggerProvider, loggerName string, minLevel slog.Leveler) *logHook {
	return &logHook{
		provider: provider,
		logger:   provider.Logger(loggerName),
		level:    minLevel.Level(),
	}
}

// Enabled reports whether the hook handles records at the given level.
func (h *logHook) Enabled(level slog.Level) bool {
	return level >= h.level
}

// Fire processes a log event and exports it to OpenTelemetry.
func (h *logHook) Fire(t time.Time, lvl slog.Level, msg []byte) error {
	// Create log record
	var logRecord log.Record
	logRecord.SetTimestamp(t)
	logRecord.SetBody(log.StringValue(string(msg)))
	logRecord.SetSeverity(mapSeverity(lvl))
	logRecord.SetSeverityText(lvl.String())

	// Emit the log record
	h.logger.Emit(context.Background(), logRecord)

	return nil
}

// mapSeverity converts slog.Level to OpenTelemetry log.Severity.
func mapSeverity(level slog.Level) log.Severity {
	switch {
	case level < slog.LevelInfo:
		return log.SeverityDebug
	case level < slog.LevelWarn:
		return log.SeverityInfo
	case level < slog.LevelError:
		return log.SeverityWarn
	case level < imgproxylogger.LevelCritical:
		return log.SeverityError
	default:
		return log.SeverityFatal
	}
}
