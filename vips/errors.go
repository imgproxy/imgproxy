package vips

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	VipsError  string
	ColorError string
)

func newVipsError(msg string) error {
	return ierrors.Wrap(VipsError(msg), 1)
}

func newVipsErrorf(format string, args ...interface{}) error {
	return ierrors.Wrap(VipsError(fmt.Sprintf(format, args...)), 1)
}

func (e VipsError) Error() string { return string(e) }

func newColorError(format string, args ...interface{}) error {
	return ierrors.Wrap(
		ColorError(fmt.Sprintf(format, args...)),
		1,
		ierrors.WithShouldReport(false),
	)
}

func (e ColorError) Error() string { return string(e) }
