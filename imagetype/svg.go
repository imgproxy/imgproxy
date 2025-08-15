package imagetype

import (
	"errors"
	"io"
	"strings"

	"github.com/imgproxy/imgproxy/v3/bufreader"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"
)

func init() {
	// Register SVG detector.
	// We register it with a priority of 100 to run it after magic number detectors
	RegisterDetector(100, IsSVG)
}

func IsSVG(r bufreader.ReadPeeker) (Type, error) {
	l := xml.NewLexer(parse.NewInput(r))

	for {
		tt, _ := l.Next()

		switch tt {
		case xml.ErrorToken:
			err := l.Err()

			if err == io.EOF || err == io.ErrUnexpectedEOF {
				// EOF or unexpected EOF means we don't have enough data to determine the type
				return Unknown, nil
			}

			var perr *parse.Error
			if errors.As(err, &perr) {
				// If the error is a parse error, we can assume that the data is not SVG
				return Unknown, nil
			}

			return Unknown, err

		case xml.StartTagToken:
			tag := strings.ToLower(string(l.Text()))
			if tag == "svg" || tag == "svg:svg" {
				return SVG, nil
			}
		}
	}
}
