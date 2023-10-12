package imagemeta

import (
	"io"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"
)

func IsSVG(r io.Reader) bool {
	maxBytes := config.MaxSvgCheckBytes

	l := xml.NewLexer(parse.NewInput(io.LimitReader(r, int64(maxBytes))))

	for {
		tt, _ := l.Next()

		switch tt {
		case xml.ErrorToken:
			return false

		case xml.StartTagToken:
			return strings.ToLower(string(l.Text())) == "svg"
		}
	}
}
