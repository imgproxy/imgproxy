package testutil

import (
	"fmt"
	"strings"
)

// OptionArgs is a helper for building option arguments in tests.
// It allows setting arguments by index and automatically trims empty arguments from the end.
// It allows setting up to 32 arguments, which should be enough for all options.
type OptionArgs struct {
	args [32]string
}

// Set sets the argument at the specified index to the string representation of the value.
func (a *OptionArgs) Set(index int, value any) *OptionArgs {
	a.args[index] = fmt.Sprintf("%v", value)
	return a
}

// String returns the string representation of the arguments, trimming empty arguments from the end.
func (a *OptionArgs) String() string {
	// Trim empty args from the end
	args := a.args[:]
	for len(args) > 0 && args[len(args)-1] == "" {
		args = args[:len(args)-1]
	}

	return strings.Join(args, ":")
}

// OptionsBuilder is a helper for building options strings in tests.
type OptionsBuilder map[string]*OptionArgs

// NewOptionsBuilder creates a new OptionsBuilder.
func NewOptionsBuilder() OptionsBuilder {
	return make(OptionsBuilder)
}

// Add adds a new option with the specified key
// and returns the OptionArgs for setting its arguments.
// If the option already exists, it returns the existing OptionArgs.
func (o OptionsBuilder) Add(key string) *OptionArgs {
	if _, exists := o[key]; !exists {
		o[key] = &OptionArgs{}
	}

	return o[key]
}

// String returns the string representation of the options, joining them with '/'.
func (o OptionsBuilder) String() string {
	formatted := make([]string, 0, len(o))

	for k, v := range o {
		formatted = append(formatted, fmt.Sprintf("%s:%s", k, v.String()))
	}

	return strings.Join(formatted, "/")
}
