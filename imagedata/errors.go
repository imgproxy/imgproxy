package imagedata

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
)

func wrapDownloadError(err error, desc string) error {
	return ierrors.Wrap(
		imagefetcher.WrapError(err), 0,
		ierrors.WithPrefix(fmt.Sprintf("can't download %s", desc)),
	)
}
