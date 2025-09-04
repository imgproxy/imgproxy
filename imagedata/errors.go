package imagedata

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

func wrapDownloadError(err error, desc string) error {
	return ierrors.Wrap(
		fetcher.WrapError(err), 0,
		ierrors.WithPrefix(fmt.Sprintf("can't download %s", desc)),
	)
}
