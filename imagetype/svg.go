package imagetype

import (
	"errors"
	"io"

	"github.com/imgproxy/imgproxy/v3/bufreader"
	xmlparser "github.com/imgproxy/imgproxy/v3/xmlparser"
)

func init() {
	// Register SVG detector.
	// We register it with a priority of 100 to run it after magic number detectors
	RegisterDetector(100, IsSVG)
}

func IsSVG(r bufreader.ReadPeeker, _, _ string) (Type, error) {
	dec := xmlparser.NewDecoder(r)

	for {
		tok, err := dec.Token()
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// EOF or unexpected EOF means we don't have enough data to determine the type
			return Unknown, nil
		}
		if err != nil {
			var perr xmlparser.SyntaxError
			if errors.As(err, &perr) {
				// If the error is a parse error, we can assume that the data is not SVG
				return Unknown, nil
			}

			return Unknown, err
		}

		if se, ok := tok.(xmlparser.StartElement); ok && se.Name.Local == "svg" {
			return SVG, nil
		}
	}
}
