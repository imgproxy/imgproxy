package env

import (
	"fmt"
	"log/slog"
	"os"
)

// Desc describes an environment variable
type Desc struct {
	Name   string
	Format string
}

// Describe creates a new EnvDesc
func Describe(name string, format string) Desc {
	return Desc{
		Name:   name,
		Format: format,
	}
}

// Getenv returns the value of the env variable
func (d Desc) Get() (string, bool) {
	value := os.Getenv(d.Name)
	return value, len(value) > 0
}

// Warn logs a warning with the env var details
func (d Desc) Warn(msg string, args ...any) {
	v, _ := d.Get()
	args = append(args, "name", d.Name, "format", d.Format, "value", v)

	slog.Warn(msg, args...)
}

// Errorf formats an error message for invalid env var
func (d Desc) Errorf(msg string, args ...any) error {
	return fmt.Errorf(
		"invalid %s value (format: %s): %s",
		d.Name,
		d.Format,
		fmt.Sprintf(msg, args...),
	)
}

// WarnParseError logs a warning when an env var fails to parse
func (d Desc) ErrorParse(err error) error {
	return d.Errorf("failed to parse: %s", err)
}

// ErrorEmpty formats an error message for empty env var
func (d Desc) ErrorEmpty() error {
	return d.Errorf("cannot be empty")
}

// ErrorRange formats an error message for out of range env var
func (d Desc) ErrorRange() error {
	return d.Errorf("out of range")
}

// ErrorZeroOrLess formats an error message for zero or less env var
func (d Desc) ErrorZeroOrNegative() error {
	return d.Errorf("cannot be zero or negative")
}

// ErrorNegative formats an error message for negative env var
func (d Desc) ErrorNegative() error {
	return d.Errorf("cannot be negative")
}
