package xmlparser

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type (
	SyntaxError struct{ *errctx.TextError }
)

func newSyntaxError(msg string, args ...any) error {
	return SyntaxError{errctx.NewTextError(
		fmt.Sprintf(msg, args...),
		1,
		errctx.WithPublicMessage("SVG syntax error"),
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
	)}
}
