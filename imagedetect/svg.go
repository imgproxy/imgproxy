package imagedetect

import (
	"strings"

	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/imgproxy/imgproxy/v3/imagetype_new"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"
)

func init() {
	// Register SVG detector (needs at least 1000 bytes to reliably detect SVG)
	RegisterDetector(IsSVG, 1000)
}

func IsSVG(r bufreader.ReadPeeker) (imagetype_new.Type, error) {
	l := xml.NewLexer(parse.NewInput(r))

	for {
		tt, _ := l.Next()

		switch tt {
		case xml.ErrorToken:
			return imagetype_new.Unknown, nil

		case xml.StartTagToken:
			tag := strings.ToLower(string(l.Text()))
			if tag == "svg" || tag == "svg:svg" {
				return imagetype_new.SVG, nil
			}
		}
	}
}
