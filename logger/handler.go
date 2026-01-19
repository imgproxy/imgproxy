package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"
)

// LevelCritical is a log level for fatal errors
const LevelCritical = slog.LevelError + 8

// Format represents the log format
type Format int

const (
	// FormatStructured is a key=value structured format
	FormatStructured Format = iota
	// FormatPretty is a human-readable format with colorization
	FormatPretty
	// FormatJSON is a JSON format
	FormatJSON
	// FormatGCP is a JSON format for Google Cloud Platform
	FormatGCP
)

// attrGroup represents a named group of attributes.
//
// Both the group name and the attributes are optional.
// Non-empty name means new nested group.
type attrGroup struct {
	name  string
	attrs []slog.Attr
}

// Hook is an interface that defines a log hook.
type Hook interface {
	// Enabled checks if the hook is enabled for the given log level.
	Enabled(lvl slog.Level) bool

	// Fire is a function that gets called on log events.
	//
	// The slice provided in the msg parameter contains the formatted log message,
	// followed by a newline character.
	// It is guaranteed to be available for the duration of the hook call.
	// The hook should not modify the contents of the msg slice except for appending.
	Fire(ctx context.Context, time time.Time, lvl slog.Level, msg []byte) error
}

// Handler is an implementation of [slog.Handler] with support for hooks.
type Handler struct {
	out    io.Writer
	config *Config

	mu *sync.Mutex // Mutex is shared between all instances

	groups []attrGroup

	hooks []Hook
}

// NewHandler creates a new [Handler] instance.
func NewHandler(out io.Writer, config *Config) *Handler {
	return &Handler{
		out:    out,
		config: config,
		mu:     new(sync.Mutex),
	}
}

// AddHook adds a new hook to the handler.
func (h *Handler) AddHook(hook Hook) {
	if hook == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.hooks = append(h.hooks, hook)
}

// Level returns the minimum log level for the handler.
func (h *Handler) Level() slog.Leveler {
	return h.config.Level
}

// Enabled checks if the given log level is enabled.
func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	if level >= h.config.Level.Level() {
		return true
	}

	for _, hook := range h.hooks {
		if hook.Enabled(level) {
			return true
		}
	}

	return false
}

// WithAttrs returns a new handler with the given attributes added.
func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return h.withGroup(attrGroup{
		name:  "",
		attrs: attrs,
	})
}

// WithGroup returns a new handler with the given group name added.
func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return h.withGroup(attrGroup{
		name:  name,
		attrs: nil,
	})
}

// Handle processes a log record.
func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	buf := newBuffer()
	defer func() {
		buf.free()
	}()

	h.format(r, buf)

	h.mu.Lock()
	defer h.mu.Unlock()

	var errs []error

	// Write log entry to output
	if r.Level >= h.config.Level.Level() {
		_, err := h.out.Write(*buf)
		if err != nil {
			errs = append(errs, err)
		}
	}

	// Fire hooks
	for _, hook := range h.hooks {
		if !hook.Enabled(r.Level) {
			continue
		}
		if err := hook.Fire(ctx, r.Time, r.Level, slices.Clip(*buf)); err != nil {
			errs = append(errs, err)
		}
	}

	// If writing to output or firing hooks returned errors,
	// join them, write to STDERR, and return
	if err := h.joinErrors(errs); err != nil {
		h.writeError(err)
		return err
	}

	return nil
}

// format formats a log record and writes it to the buffer.
func (h *Handler) format(r slog.Record, buf *buffer) {
	groups := h.groups

	// If there are no attributes in the record itself,
	// remove empty groups from the end
	if r.NumAttrs() == 0 {
		for len(groups) > 0 && len(groups[len(groups)-1].attrs) == 0 {
			groups = groups[:len(groups)-1]
		}
	}

	// Format the log record according to the format specified in options
	switch h.config.Format {
	case FormatPretty:
		newFormatterPretty(groups, buf).format(r)
	case FormatJSON:
		newFormatterJSON(groups, buf, false).format(r)
	case FormatGCP:
		newFormatterJSON(groups, buf, true).format(r)
	default:
		newFormatterStructured(groups, buf).format(r)
	}

	// Add line break after each log entry
	buf.append('\n')
}

func (h *Handler) joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return errors.Join(errs...)
}

// writeError writes a logging error message to STDERR.
func (h *Handler) writeError(err error) {
	buf := newBuffer()
	defer func() {
		buf.free()
	}()

	r := slog.NewRecord(time.Now(), slog.LevelError, "An error occurred during logging", 0)
	r.Add("error", err)

	h.format(r, buf)

	_, _ = os.Stderr.Write(*buf)
}

// withGroup returns a new handler with the given attribute group added.
func (h *Handler) withGroup(group attrGroup) *Handler {
	h2 := *h
	h2.groups = append(slices.Clip(h.groups), group)
	h2.hooks = slices.Clip(h.hooks)
	return &h2
}
