package color

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type ColorError struct{ *errctx.TextError }

func newColorError(format string, args ...any) error {
	return ColorError{errctx.NewTextError(
		fmt.Sprintf(format, args...),
		1,
		errctx.WithShouldReport(false),
	)}
}
