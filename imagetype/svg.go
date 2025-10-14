package imagetype

import (
	"encoding/xml"
	"errors"
	"io"

	"github.com/imgproxy/imgproxy/v3/bufreader"
)

func init() {
	// Register SVG detector.
	// We register it with a priority of 100 to run it after magic number detectors
	RegisterDetector(100, IsSVG)
}

func IsSVG(r bufreader.ReadPeeker) (Type, error) {
	dec := xml.NewDecoder(r)
	dec.Strict = false

	for {
		tok, err := dec.RawToken()
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// EOF or unexpected EOF means we don't have enough data to determine the type
			return Unknown, nil
		}
		if err != nil {
			var perr *xml.SyntaxError
			if errors.As(err, &perr) {
				// If the error is a parse error, we can assume that the data is not SVG
				return Unknown, nil
			}

			return Unknown, err
		}

		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "svg" {
			return SVG, nil
		}
	}
}
