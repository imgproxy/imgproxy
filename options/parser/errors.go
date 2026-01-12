package optionsparser

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

const processingDocsUrl = "https://docs.imgproxy.net/usage/processing"

type (
	InvalidURLError      struct{ *errctx.TextError }
	UnknownOptionError   struct{ *errctx.TextError }
	ForbiddenOptionError struct{ *errctx.TextError }
	OptionArgumentError  struct{ *errctx.TextError }
	SecurityOptionsError struct{ *errctx.TextError }
)

func newInvalidURLError(format string, args ...any) error {
	return InvalidURLError{errctx.NewTextError(
		fmt.Sprintf(format, args...),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithDocsURL(processingDocsUrl),
		errctx.WithShouldReport(false),
	)}
}

func newUnknownOptionError(kind, opt string) error {
	return UnknownOptionError{errctx.NewTextError(
		fmt.Sprintf("Unknown %s option %s", kind, opt),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithDocsURL(processingDocsUrl),
		errctx.WithShouldReport(false),
	)}
}

func newForbiddenOptionError(kind, opt string) error {
	return ForbiddenOptionError{errctx.NewTextError(
		fmt.Sprintf("Forbidden %s option %s", kind, opt),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithShouldReport(false),
		errctx.WithDocsURL(processingDocsUrl),
	)}
}

func newOptionArgumentError(key, format string, args ...any) error {
	url := processingDocsUrl
	if key != "" {
		url = url + "#" + errorKey(key)
	}

	return OptionArgumentError{errctx.NewTextError(
		fmt.Sprintf(format, args...),
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithDocsURL(url),
		errctx.WithShouldReport(false),
	)}
}

func newSecurityOptionsError() error {
	return SecurityOptionsError{errctx.NewTextError(
		"Security processing options are not allowed",
		1,
		errctx.WithStatusCode(http.StatusForbidden),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithDocsURL(processingDocsUrl),
		errctx.WithShouldReport(false),
	)}
}

// newInvalidArgsError creates a standardized error for invalid arguments
func newInvalidArgsError(key string, args []string) error {
	return newOptionArgumentError(key, "Invalid %s arguments: %s", key, args)
}

// newInvalidArgumentError creates a standardized error for an invalid single argument
func newInvalidArgumentError(key, arg string, expected ...string) error {
	msg := "Invalid %s: %s"
	if len(expected) > 0 {
		msg += " (expected " + strings.Join(expected, ", ") + ")"
	}

	return newOptionArgumentError(key, msg, key, arg)
}

func errorKey(key string) string {
	k, _, _ := strings.Cut(strings.ReplaceAll(key, "_", "-"), ".")
	return k
}
