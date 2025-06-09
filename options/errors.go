package options

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	InvalidURLError     string
	UnknownOptionError  string
	OptionArgumentError string
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
