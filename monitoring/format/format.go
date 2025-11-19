package format

import (
	"fmt"
	"strings"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

// FormatSegmentName formats segment name string
func FormatSegmentName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "_")
}

// FormatErrType formats error type string with error concrete type
func FormatErrType(errType string, err error) string {
	return fmt.Sprintf("%s_error (%s)", errType, errctx.ErrorType(err))
}
