package otel

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	sdklog "go.opentelemetry.io/otel/sdk/log"

	"github.com/imgproxy/imgproxy/v4/logger"
)

// logHook is a logger.Hook that exports logs to OpenTelemetry.
type logHook struct {
	provider *sdklog.LoggerProvider
	handler  slog.Handler
	level    slog.Level
}

// newLogHook creates a new OpenTelemetry log hook.
func newLogHook(provider *sdklog.LoggerProvider, loggerName string, minLevel slog.Leveler) *logHook {
	handler := otelslog.NewHandler(
		loggerName,
		otelslog.WithLoggerProvider(provider),
	)

	return &logHook{
		provider: provider,
		handler:  handler,
		level:    minLevel.Level(),
	}
}

// Enabled reports whether the hook handles records at the given level.
func (h *logHook) Enabled(level slog.Level) bool {
	// We use [sdklog.BatchProcessor] in the OpenTelemetry logger provider,
	// which accepts all log records no matter their severity,
	// so we can't rely on [h.handler.Enabled] here.
	// Instead, we check the log level against the minimum level configured for the hook.
	return level >= h.level
}

// Fire processes a log event and exports it to OpenTelemetry.
func (h *logHook) Fire(ctx context.Context, r slog.Record, groups []slog.Attr, msg []byte) error {
	handler := h.handler

	// Apply groups and attrs added with [Handler.WithAttrs] and [Handler.WithGroup]
	logger.ProcessGroups(
		groups,
		func(name string) { handler = handler.WithGroup(name) },
		func(attrs []slog.Attr) { handler = handler.WithAttrs(attrs) },
	)

	// Emit the log record
	if err := handler.Handle(ctx, r); err != nil {
		return err
	}

	// Ensure logs are flushed for critical errors
	if r.Level >= logger.LevelCritical {
		flushCtx, cancel := context.WithTimeout(ctx, stopTimeout)
		defer cancel()

		return h.provider.ForceFlush(flushCtx)
	}

	return nil
}
