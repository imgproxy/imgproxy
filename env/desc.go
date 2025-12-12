package env

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

const (
	docsUrl = "https://docs.imgproxy.net/configuration/options#"
)

// ParseFn is a function type that defines a parser for a specific type V
type ParseFn[V any] func(string) (V, error)

// Desc describes an environment variable with a specific type and parser
type Desc[V any, F ParseFn[V]] struct {
	Name            string
	format          string
	docsUrlOverride string
	parseFn         F
}

// GetEnv returns the value of the env variable
func (d *Desc[V, F]) GetEnv() (string, bool) {
	if len(d.Name) == 0 {
		return "", false
	}

	// Fall back to environment variable
	value := os.Getenv(d.Name)
	return value, len(value) > 0
}

// Parse parses the environment variable and sets the value
func (d *Desc[V, F]) Parse(value *V) error {
	env, ok := d.GetEnv()
	if !ok || strings.TrimSpace(env) == "" {
		return nil
	}

	parsedValue, err := d.parseFn(env)
	if err != nil {
		return d.Errorf("parse error: %w", err)
	}

	*value = parsedValue
	return nil
}

// WithDocsURL sets a custom documentation URL for the env var
func (d Desc[V, F]) WithDocsURL(url string) Desc[V, F] {
	d.docsUrlOverride = url
	return d
}

// WithFormat sets a custom format description for the env var
func (d Desc[V, F]) WithFormat(format string) Desc[V, F] {
	d.format = format
	return d
}

// ErrorParse logs a warning when an env var fails to parse
func (d *Desc[V, F]) ErrorParse(err error) error {
	return d.Errorf("failed to parse: %w", err)
}

// ErrorEmpty formats an error message for empty env var
func (d *Desc[V, F]) ErrorEmpty() error {
	return d.Errorf("cannot be empty")
}

// ErrorRange formats an error message for out of range env var
func (d *Desc[V, F]) ErrorRange() error {
	return d.Errorf("out of range")
}

// ErrorZeroOrNegative formats an error message for zero or less env var
func (d *Desc[V, F]) ErrorZeroOrNegative() error {
	return d.Errorf("cannot be zero or negative")
}

// ErrorNegative formats an error message for negative env var
func (d *Desc[V, F]) ErrorNegative() error {
	return d.Errorf("cannot be negative")
}

// Warn logs a warning with the env var details
func (d *Desc[V, F]) Warn(msg string, args ...any) {
	v, _ := d.GetEnv()
	args = append(args, "name", d.Name, "format", d.format, "value", v)

	slog.Warn(msg, args...)
}

// Errorf formats an error message for invalid env var
func (d *Desc[V, F]) Errorf(msg string, args ...any) error {
	return fmt.Errorf(
		"invalid %s value (format: %s): %w, see %s",
		d.Name,
		d.format,
		fmt.Errorf(msg, args...),
		d.docsUrl(),
	)
}

// docsUrl returns the documentation URL for the env var
func (d *Desc[V, F]) docsUrl() string {
	if len(d.docsUrlOverride) > 0 {
		return d.docsUrlOverride
	}

	return docsUrl + d.Name
}
