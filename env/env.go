package env

import (
	"fmt"
	"os"
)

// ParseFunc is a function type for parsing environment variable values
type ParseFunc[T any] func(string) (T, error)

// ValidateFunc is a function type for validating parsed environment variable values
type ValidateFunc[T any] func(T) bool

// generic number (add more types if needed)
type number = interface {
	~int | ~int64 | ~float32 | ~float64
}

// EnvVar represents an environment variable with its metadata
type EnvVar[T any] struct {
	name        string            // ENV variable name
	description string            // Description for help message
	formatDesc  string            // Description of the format
	def         T                 // Default value (if any)
	parse       ParseFunc[T]      // Function to parse the value from string
	validate    []ValidateFunc[T] // Optional validation function(s)
}

// Define creates a new EnvVar with the given parameters
func Define[T any](name, description, formatDesc string, parse ParseFunc[T], def T, val ...ValidateFunc[T]) EnvVar[T] {
	return EnvVar[T]{
		name:        name,
		description: description,
		parse:       parse,
		formatDesc:  formatDesc,
		def:         def,
		validate:    val,
	}
}

// Get retrieves the value of the environment variable, parses it,
// validates it, and sets it to the provided pointer.
func (e EnvVar[T]) Get(value *T) (err error) {
	val, exists := os.LookupEnv(e.name)

	// First, let's fill with a value (either default or parsed)
	if !exists {
		*value = e.def
		return nil
	} else {
		*value, err = e.parse(val)
		if err != nil {
			return fmt.Errorf("error parsing %s (required: %v): %v", e.name, e.formatDesc, err)
		}
	}

	// Then, validate the value
	for _, validate := range e.validate {
		if !validate(*value) {
			return fmt.Errorf("invalid value for %s (required: %v): %v", e.name, e.formatDesc, err)
		}
	}

	return nil
}

// Default returns the default value of the environment variable
func (e EnvVar[T]) Default() T {
	return e.def
}

// NotEmpty is a validation function that checks if a string is not empty
func NotEmpty(v string) bool {
	return v != ""
}

// Positive is a validation function that checks if a number is greater than zero
func Positive[T number](v T) bool {
	return v > 0
}
