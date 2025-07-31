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
)

var svgBufPool = bufpool.New("svg", config.Workers, config.DownloadBufferSize)

func BorrowBuffer() (*bytes.Buffer, context.CancelFunc) {
	buf := svgBufPool.Get(0, false)
	cancel := func() { svgBufPool.Put(buf) }

	return buf, cancel
}

func cloneHeaders(src map[string]string) map[string]string {
	if src == nil {
		return nil
	}

	dst := make(map[string]string, len(src))
	for k, v := range src {
		dst[k] = v
	}

	return dst
}

func Sanitize(data *imagedata.ImageData) (*imagedata.ImageData, error) {
	r := bytes.NewReader(data.Data)
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

			newData := imagedata.ImageData{
				Data:    buf.Bytes(),
				Type:    data.Type,
				Headers: cloneHeaders(data.Headers),
			}
			newData.SetCancel(cancel)

			return &newData, nil
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
