package options

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	TypeMismatchError struct{ error }
)

func newTypeMismatchError(key string, exp, got any) error {
	return ierrors.Wrap(
		TypeMismatchError{fmt.Errorf("option %s is %T, not %T", key, exp, got)},
		1,
	)
}
