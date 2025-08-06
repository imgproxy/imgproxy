package imagedata

import (
	"bytes"
	"context"
	"io"

	"github.com/imgproxy/imgproxy/v3/bufpool"
	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/imagemeta"
	"github.com/imgproxy/imgproxy/v3/security"
)

var downloadBufPool *bufpool.Pool

func initRead() {
	downloadBufPool = bufpool.New("download", config.Workers, config.DownloadBufferSize)
}

func readAndCheckImage(r io.Reader, contentLength int, secopts security.Options) (ImageData, error) {
	buf := downloadBufPool.Get(contentLength, false)
	cancel := func() { downloadBufPool.Put(buf) }

	br := bufreader.New(r, buf)

	meta, err := imagemeta.DecodeMeta(br)
	if err != nil {
		buf.Reset()
		cancel()

		return nil, imagefetcher.WrapError(err)
	}

	if err = security.CheckDimensions(meta.Width(), meta.Height(), 1, secopts); err != nil {
		buf.Reset()
		cancel()

		return nil, imagefetcher.WrapError(err)
	}

	downloadBufPool.GrowBuffer(buf, contentLength)

	if err = br.Flush(); err != nil {
		buf.Reset()
		cancel()

		return nil, imagefetcher.WrapError(err)
	}

	i := NewFromBytesWithFormat(meta.Format(), buf.Bytes())
	i.AddCancel(cancel)
	return i, nil
}

func BorrowBuffer() (*bytes.Buffer, context.CancelFunc) {
	buf := downloadBufPool.Get(0, false)
	cancel := func() { downloadBufPool.Put(buf) }

	return buf, cancel
}
