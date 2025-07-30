package imagedata

import (
	"context"
	"sync"

	"github.com/imgproxy/imgproxy/v3/imagetype"
)

type ImageData struct {
	Type    imagetype.Type
	Data    []byte
	Headers map[string]string

	cancel     context.CancelFunc
	cancelOnce sync.Once
}

func (d *ImageData) Close() {
	d.cancelOnce.Do(func() {
		if d.cancel != nil {
			d.cancel()
		}
	})
}

func (d *ImageData) SetCancel(cancel context.CancelFunc) {
	d.cancel = cancel
}

// func Init() error {
// 	if err := initDownloading(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func Download(ctx context.Context, imageURL, desc string, opts DownloadOptions, secopts security.Options) (*ImageData, error) {
// 	imgdata, err := download(ctx, imageURL, opts, secopts)
// 	if err != nil {
// 		return nil, ierrors.Wrap(
// 			err, 0,
// 			ierrors.WithPrefix(fmt.Sprintf("Can't download %s", desc)),
// 		)
// 	}

// 	return From(imgdata), nil
// }
