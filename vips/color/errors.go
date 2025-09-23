package color

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type ColorError string

func newColorError(format string, args ...interface{}) error {
	return ierrors.Wrap(
		ColorError(fmt.Sprintf(format, args...)),
		1,
		ierrors.WithShouldReport(false),
	)
}

func (e ColorError) Error() string { return string(e) }
