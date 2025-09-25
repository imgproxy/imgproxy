package optionsparser

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	InvalidURLError      string
	UnknownOptionError   string
	OptionArgumentError  string
	SecurityOptionsError struct{}
)

func newInvalidURLError(format string, args ...interface{}) error {
	return ierrors.Wrap(
		InvalidURLError(fmt.Sprintf(format, args...)),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e InvalidURLError) Error() string { return string(e) }

func newUnknownOptionError(kind, opt string) error {
	return ierrors.Wrap(
		UnknownOptionError(fmt.Sprintf("Unknown %s option %s", kind, opt)),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func newForbiddenOptionError(kind, opt string) error {
	return ierrors.Wrap(
		UnknownOptionError(fmt.Sprintf("Forbidden %s option %s", kind, opt)),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e UnknownOptionError) Error() string { return string(e) }

func newOptionArgumentError(format string, args ...interface{}) error {
	return ierrors.Wrap(
		OptionArgumentError(fmt.Sprintf(format, args...)),
		1,
		ierrors.WithStatusCode(http.StatusNotFound),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e OptionArgumentError) Error() string { return string(e) }

func newSecurityOptionsError() error {
	return ierrors.Wrap(
		SecurityOptionsError{},
		1,
		ierrors.WithStatusCode(http.StatusForbidden),
		ierrors.WithPublicMessage("Invalid URL"),
		ierrors.WithShouldReport(false),
	)
}

func (e SecurityOptionsError) Error() string { return "Security processing options are not allowed" }

// newInvalidArgsError creates a standardized error for invalid arguments
func newInvalidArgsError(name string, args []string) error {
	return newOptionArgumentError("Invalid %s arguments: %s", name, args)
}

// newInvalidArgumentError creates a standardized error for an invalid single argument
func newInvalidArgumentError(key, arg string, expected ...string) error {
	msg := "Invalid %s: %s"
	if len(expected) > 0 {
		msg += " (expected " + strings.Join(expected, ", ") + ")"
	}

	return newOptionArgumentError(msg, key, arg)
}
