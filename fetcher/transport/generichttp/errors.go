package generichttp

import (
	"net/http"

	"github.com/imgproxy/imgproxy/v4/errctx"
)

type (
	SourceAddressError struct{ *errctx.TextError }
)

func newSourceAddressError(msg string) error {
	return SourceAddressError{errctx.NewTextError(
		msg,
		1,
		errctx.WithStatusCode(http.StatusNotFound),
		errctx.WithPublicMessage("Invalid source URL"),
		errctx.WithShouldReport(false),
	)}
}
