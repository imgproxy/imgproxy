package svg

import (
	"bytes"
	"errors"
	"sync"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/xmlparser"
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

	doc, err := xmlparser.NewDocument(data.Reader())
	if err != nil {
		return nil, newSanitizeError(err)
	}

	// Sanitize the document's children
	p.sanitizeChildren(&doc.Node)

	buf, ok := pool.Get().(*bytes.Buffer)
	if !ok {
		return nil, newSanitizeError(errors.New("svg.Sanitize: failed to get buffer from pool"))
	}
	buf.Reset()

	cancel := func() {
		pool.Put(buf)
	}

	// Write the sanitized document to the buffer
	if _, err := doc.WriteTo(buf); err != nil {
		cancel()
		return nil, newSanitizeError(err)
	}

	// Create new ImageData from the sanitized buffer
	newData := imagedata.NewFromBytesWithFormat(
		imagetype.SVG,
		buf.Bytes(),
	)
	newData.AddCancel(cancel)

	return newData, nil
}

// sanitizeChildren sanitizes all child elements of the given element.
func (p *Processor) sanitizeChildren(el *xmlparser.Node) {
	if el == nil || len(el.Children) == 0 {
		return
	}

	// Filter children in place
	filteredChildren := el.Children[:0]
	for _, toc := range el.Children {
		childEl, ok := toc.(*xmlparser.Node)
		if !ok {
			// Keep non-element nodes (text, comments, etc.)
			filteredChildren = append(filteredChildren, toc)
			continue
		}

		// Sanitize the child element.
		// Keep this child if sanitizeElement returned true.
		if p.sanitizeElement(childEl) {
			filteredChildren = append(filteredChildren, childEl)
		}
	}

	el.Children = filteredChildren
}

// sanitizeElement sanitizes a single SVG element.
// It returns true if the element should be kept, false if it should be removed.
func (p *Processor) sanitizeElement(el *xmlparser.Node) bool {
	if el == nil {
		return false
	}

	// Strip <script> tags
	if el.Name.Local() == "script" {
		return false
	}

	// Filter out unsafe attributes (such as on* events)
	el.Attrs.Filter(func(attr *xmlparser.Attribute) bool {
		_, unsafe := unsafeAttrs[attr.Name.Local()]
		return !unsafe
	})

	// Special handling for <use> tags.
	if el.Name.Local() == "use" {
		el.Attrs.Filter(func(attr *xmlparser.Attribute) bool {
			// Keep non-href attributes
			if attr.Name.Local() != "href" {
				return true
			}
			// Strip hrefs that are not internal references
			return len(attr.Value) == 0 || attr.Value[0] == '#'
		})
	}

	// Recurse into children
	p.sanitizeChildren(el)

	// Keep this element
	return true
}
