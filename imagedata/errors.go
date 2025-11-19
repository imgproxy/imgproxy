package imagedata

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/fetcher"
)

type FileSizeError struct{ *errctx.TextError }

func newFileSizeError() error {
	return FileSizeError{errctx.NewTextError(
		"Source image file is too big",
		1,
		errctx.WithStatusCode(http.StatusUnprocessableEntity),
		errctx.WithPublicMessage("Invalid source image"),
		errctx.WithShouldReport(false),
	)}
}

func wrapDownloadError(err error, desc string) error {
	return errctx.Wrap(
		fetcher.WrapError(err, 1),
		errctx.WithPrefix(fmt.Sprintf("can't download %s", desc)),
	)
}
