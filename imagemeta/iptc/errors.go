package iptc

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
)

type IptcError struct{ *errctx.TextError }

func newIptcError(format string, args ...interface{}) error {
	return IptcError{errctx.NewTextError(
		fmt.Sprintf(format, args...),
		1,
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
		errctx.WithPublicMessage("Invalid IPTC data"),
		errctx.WithShouldReport(false),
	)}
}
