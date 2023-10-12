package imagedata

import (
	"bytes"
	"context"
	"io"

	"github.com/imgproxy/imgproxy/v3/bufpool"
	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/security"
)

var ErrSourceImageTypeNotSupported = ierrors.New(422, "Source image type not supported", "Invalid source image")

var downloadBufPool *bufpool.Pool

func initRead() {
	downloadBufPool = bufpool.New("download", config.Workers, config.DownloadBufferSize)
}

func readAndCheckImage(r io.Reader, contentLength int, secopts security.Options) (*ImageData, error) {
	if err := security.CheckFileSize(contentLength, secopts); err != nil {
		return nil, err
	}

	buf := downloadBufPool.Get(contentLength, false)
	cancel := func() { downloadBufPool.Put(buf) }

	r = security.LimitFileSize(r, secopts)

	br := bufreader.New(r, buf)

	meta, err := imagemeta.DecodeMeta(br)
	if err != nil {
		buf.Reset()
		cancel()

		if err == imagemeta.ErrFormat {
			return nil, ErrSourceImageTypeNotSupported
		}

		return nil, wrapError(err)
	}

	if err = security.CheckDimensions(meta.Width(), meta.Height(), 1, secopts); err != nil {
		buf.Reset()
		cancel()

		return nil, wrapError(err)
	}

	downloadBufPool.GrowBuffer(buf, contentLength)

	if err = br.Flush(); err != nil {
		buf.Reset()
		cancel()

		return nil, wrapError(err)
	}

	return &ImageData{
		Data:   buf.Bytes(),
		Type:   meta.Format(),
		cancel: cancel,
	}, nil
}

func BorrowBuffer() (*bytes.Buffer, context.CancelFunc) {
	buf := downloadBufPool.Get(0, false)
	cancel := func() { downloadBufPool.Put(buf) }

	return buf, cancel
}
