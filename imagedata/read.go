package imagedata

import (
	"io"

	"github.com/imgproxy/imgproxy/v3/bufpool"
	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/security"
)

var (
	ErrSourceFileTooBig            = ierrors.New(422, "Source image file is too big", "Invalid source image")
	ErrSourceImageTypeNotSupported = ierrors.New(422, "Source image type not supported", "Invalid source image")
)

var downloadBufPool *bufpool.Pool

func initRead() {
	downloadBufPool = bufpool.New("download", config.Concurrency, config.DownloadBufferSize)
}

type hardLimitReader struct {
	r    io.Reader
	left int
}

func (lr *hardLimitReader) Read(p []byte) (n int, err error) {
	if lr.left <= 0 {
		return 0, ErrSourceFileTooBig
	}
	if len(p) > lr.left {
		p = p[0:lr.left]
	}
	n, err = lr.r.Read(p)
	lr.left -= n
	return
}

func readAndCheckImage(r io.Reader, contentLength int) (*ImageData, error) {
	if config.MaxSrcFileSize > 0 && contentLength > config.MaxSrcFileSize {
		return nil, ErrSourceFileTooBig
	}

	buf := downloadBufPool.Get(contentLength)
	cancel := func() { downloadBufPool.Put(buf) }

	if config.MaxSrcFileSize > 0 {
		r = &hardLimitReader{r: r, left: config.MaxSrcFileSize}
	}

	br := bufreader.New(r, buf)

	meta, err := imagemeta.DecodeMeta(br)
	if err == imagemeta.ErrFormat {
		return nil, ErrSourceImageTypeNotSupported
	}
	if err != nil {
		return nil, checkTimeoutErr(err)
	}

	if err = security.CheckDimensions(meta.Width(), meta.Height()); err != nil {
		return nil, err
	}

	if err = br.Flush(); err != nil {
		cancel()
		return nil, checkTimeoutErr(err)
	}

	return &ImageData{
		Data:   buf.Bytes(),
		Type:   meta.Format(),
		cancel: cancel,
	}, nil
}
