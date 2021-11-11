package imagemeta

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"
	"golang.org/x/text/encoding/charmap"
)

type svgHeader struct {
	XMLName xml.Name
}

func xmlCharsetReader(charset string, input io.Reader) (io.Reader, error) {
	if strings.EqualFold(charset, "iso-8859-1") {
		return charmap.ISO8859_1.NewDecoder().Reader(input), nil
	}
	return nil, fmt.Errorf("Unknown SVG charset: %s", charset)
}

func IsSVG(r io.Reader) (bool, error) {
	maxBytes := config.MaxSvgCheckBytes

	var h svgHeader

	buf := make([]byte, 0, maxBytes)
	b := make([]byte, 1024)

	rr := bytes.NewReader(buf)

	for {
		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			return false, err
		}
		if n <= 0 {
			return false, nil
		}

		buf = append(buf, b[:n]...)
		rr.Reset(buf)

		dec := xml.NewDecoder(rr)
		dec.Strict = false
		dec.CharsetReader = xmlCharsetReader
		if dec.Decode(&h); h.XMLName.Local == "svg" {
			return true, nil
		}

		if len(buf) >= maxBytes {
			break
		}
	}

	return false, nil
}
