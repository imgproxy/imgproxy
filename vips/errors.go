package vips

import (
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type VipsError string

func newVipsError(msg string) error {
	return ierrors.Wrap(VipsError(msg), 2)
}

func (e VipsError) Error() string { return string(e) }
