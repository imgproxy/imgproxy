package options

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type (
	TypeMismatchError struct{ *errctx.TextError }
)

func newTypeMismatchError(key string, exp, got any) error {
	return TypeMismatchError{errctx.NewTextError(
		fmt.Sprintf("option %s is %T, not %T", key, exp, got),
		1,
	)}
}
