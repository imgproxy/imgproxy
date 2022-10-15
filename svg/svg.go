package svg

import (
	"bytes"
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"

	"github.com/imgproxy/imgproxy/v3/imagedata"
)

func Satitize(data *imagedata.ImageData) (*imagedata.ImageData, error) {
	r := bytes.NewReader(data.Data)
	l := xml.NewLexer(parse.NewInput(r))

	buf, cancel := imagedata.BorrowBuffer()

	ignoreTag := 0

	for {
		tt, tdata := l.Next()

		if ignoreTag > 0 {
			switch tt {
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
				Data: buf.Bytes(),
				Type: data.Type,
			}
			newData.SetCancel(cancel)

			return &newData, nil
		case xml.StartTagToken:
			if strings.ToLower(string(l.Text())) == "script" {
				ignoreTag++
				continue
			}
			buf.Write(tdata)
		case xml.AttributeToken:
			if strings.ToLower(string(l.Text())) == "onload" {
				continue
			}
			buf.Write(tdata)
		default:
			buf.Write(tdata)
		}
	}
}
