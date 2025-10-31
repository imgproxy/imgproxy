package format

import (
	"fmt"
	"strings"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

// FormatSegmentName formats segment name string
func FormatSegmentName(name string) string {
	return strings.ReplaceAll(strings.ToLower(name), " ", "_")
}

// FormatErrType formats error type string with error concrete type
func FormatErrType(errType string, err error) string {
	errType += "_error"

	if ierr, ok := err.(*ierrors.Error); ok {
		err = ierr.Unwrap()
	}

	return fmt.Sprintf("%s (%T)", errType, err)
}
