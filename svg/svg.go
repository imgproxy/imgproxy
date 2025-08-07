package svg

import (
	"bytes"
	"context"
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"

	"github.com/imgproxy/imgproxy/v3/bufpool"
	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
)

var downloadBufPool *bufpool.Pool = bufpool.New("download", config.Workers, config.DownloadBufferSize)

func BorrowBuffer() (*bytes.Buffer, context.CancelFunc) {
	buf := downloadBufPool.Get(0, false)
	cancel := func() { downloadBufPool.Put(buf) }

	return buf, cancel
}

func Sanitize(data imagedata.ImageData) (imagedata.ImageData, error) {
	r := data.Reader()
	l := xml.NewLexer(parse.NewInput(r))

	buf, cancel := BorrowBuffer()

	ignoreTag := 0

	var curTagName string

	for {
		tt, tdata := l.Next()

		if ignoreTag > 0 {
			switch tt {
			case xml.ErrorToken:
				cancel()
				return nil, l.Err()
			case xml.EndTagToken, xml.StartTagCloseVoidToken:
				ignoreTag--
			case xml.StartTagToken:
				ignoreTag++
			}

			continue
		}

		switch tt {
		case xml.ErrorToken:
			if l.Err() != io.EOF {
				cancel()
				return nil, l.Err()
			}

			newData := imagedata.NewFromBytesWithFormat(
				imagetype.SVG,
				buf.Bytes(),
			)
			newData.AddCancel(cancel)

			return newData, nil
		case xml.StartTagToken:
			curTagName = strings.ToLower(string(l.Text()))

			if curTagName == "script" {
				ignoreTag++
				continue
			}

			buf.Write(tdata)
		case xml.AttributeToken:
			attrName := strings.ToLower(string(l.Text()))

			if _, unsafe := unsafeAttrs[attrName]; unsafe {
				continue
			}

			if curTagName == "use" && (attrName == "href" || attrName == "xlink:href") {
				val := strings.TrimSpace(strings.Trim(string(l.AttrVal()), `"'`))
				if len(val) > 0 && val[0] != '#' {
					continue
				}
			}

			buf.Write(tdata)
		default:
			buf.Write(tdata)
		}
	}
}
