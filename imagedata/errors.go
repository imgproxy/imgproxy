package imagedata

import (
	"fmt"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type FileSizeError struct{}

func newFileSizeError() error {
	return ierrors.Wrap(
		FileSizeError{},
		1,
		ierrors.WithStatusCode(http.StatusUnprocessableEntity),
		ierrors.WithPublicMessage("Invalid source image"),
		ierrors.WithShouldReport(false),
	)
}

func (e FileSizeError) Error() string { return "Source image file is too big" }

func wrapDownloadError(err error, desc string) error {
	return ierrors.Wrap(
		fetcher.WrapError(err), 0,
		ierrors.WithPrefix(fmt.Sprintf("can't download %s", desc)),
	)
}
