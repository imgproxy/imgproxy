package imagemeta

import (
	"encoding/xml"
	"io"
	"sync/atomic"
)

var maxSvgBytes int64 = 32 * 1024

type svgHeader struct {
	XMLName xml.Name
}

func SetMaxSvgCheckRead(n int) {
	atomic.StoreInt64(&maxSvgBytes, int64(n))
}

func IsSVG(r io.Reader) (bool, error) {
	maxBytes := int(atomic.LoadInt64(&maxSvgBytes))

	var h svgHeader

	buf := make([]byte, 0, maxBytes)
	b := make([]byte, 1024)

	for {
		n, err := r.Read(b)
		if err != nil && err != io.EOF {
			return false, err
		}
		if n <= 0 {
			return false, nil
		}

		buf = append(buf, b[:n]...)

		if xml.Unmarshal(buf, &h); h.XMLName.Local == "svg" {
			return true, nil
		}

		if len(buf) >= maxBytes {
			break
		}
	}

	return false, nil
}
