package svg

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"sync"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/xml"
)

// pool represents temorary pool for svg sanitized data
var pool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(nil)
	},
}

// Processor provides SVG processing capabilities
type Processor struct {
	config *Config
}

// New creates a new SVG processor instance
func New(config *Config) *Processor {
	return &Processor{
		config: config,
	}
}

// Process processes the given image data
func (p *Processor) Process(o *options.Options, data imagedata.ImageData) (imagedata.ImageData, error) {
	if data.Format() != imagetype.SVG {
		return data, nil
	}

	var err error

	data, err = p.sanitize(data)
	if err != nil {
		return data, err
	}

	return data, nil
}

// sanitize sanitizes the SVG data.
// It strips <script> and unsafe attributes (on* events).
func (p *Processor) sanitize(data imagedata.ImageData) (imagedata.ImageData, error) {
	if !p.config.Sanitize {
		return data, nil
	}

	r := data.Reader()
	l := xml.NewLexer(parse.NewInput(r))

	buf, ok := pool.Get().(*bytes.Buffer)
	if !ok {
		return nil, newSanitizeError(errors.New("svg.Sanitize: failed to get buffer from pool"))
	}
	buf.Reset()

	cancel := func() {
		pool.Put(buf)
	}

	ignoreTag := 0

	var curTagName string

	for {
		tt, tdata := l.Next()

		if tt == xml.ErrorToken {
			if l.Err() != io.EOF {
				cancel()
				return nil, newSanitizeError(l.Err())
			}
			break
		}

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

	newData := imagedata.NewFromBytesWithFormat(
		imagetype.SVG,
		buf.Bytes(),
	)
	newData.AddCancel(cancel)

	return newData, nil
}
