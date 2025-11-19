package processing

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

type (
	SaveFormatError struct{ *errctx.TextError }
)

func newSaveFormatError(format imagetype.Type) error {
	return SaveFormatError{errctx.NewTextError(
		fmt.Sprintf("Can't save %s, probably not supported by your libvips", format),
		1,
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
		errctx.WithPublicMessage("Invalid URL"),
		errctx.WithShouldReport(false),
	)}
}
