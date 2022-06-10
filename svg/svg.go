package svg

import (
	"bytes"
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"
)

func Satitize(data []byte) ([]byte, error) {
	r := bytes.NewReader(data)
	l := xml.NewLexer(parse.NewInput(r))

	buf := new(bytes.Buffer)
	buf.Grow(len(data))

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
				return nil, l.Err()
			}
			return buf.Bytes(), nil
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
