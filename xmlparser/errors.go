package xmlparser

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type (
	SyntaxError struct{ error }
)

func newSyntaxError(msg string, args ...any) error {
	return ierrors.Wrap(
		SyntaxError{fmt.Errorf(msg, args...)},
		1,
		ierrors.WithPublicMessage("SVG syntax error"),
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
	)
}
