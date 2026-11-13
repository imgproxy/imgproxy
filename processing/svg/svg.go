package svg

import (
	"bytes"
	"errors"
	"strings"
	"sync"

	"github.com/imgproxy/imgproxy/v4/imagedata"
	"github.com/imgproxy/imgproxy/v4/imagetype"
	"github.com/imgproxy/imgproxy/v4/options"
	"github.com/imgproxy/imgproxy/v4/xmlparser"
)

const (
	href = "href"

	svgNS   = "http://www.w3.org/2000/svg"
	xlinkNS = "http://www.w3.org/1999/xlink"
)

// animationTargetAttrs lists attributes that SMIL animations must not target,
// since animating them can turn a safe link into a javascript: one.
var animationTargetAttrs = map[string]struct{}{
	href:  {},
	"src": {},
}

// pool represents temorary pool for svg sanitized data
var pool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(nil)
	},
}

// Processor provides SVG processing capabilities
type Processor struct {
	config *Config

	// tags and attrs are the effective allowlists (lowercase names)
	tags  map[string]struct{}
	attrs map[string]struct{}
}

// New creates a new SVG processor instance
func New(config *Config) *Processor {
	p := &Processor{
		config: config,
		tags:   defaultTagSet,
		attrs:  defaultAttrSet,
	}

	// Invalid entries are rejected by Config.Validate before we get here
	if len(config.Tags) > 0 {
		p.tags, _ = applyListOverrides(defaultTags, config.Tags)
	}

	if len(config.Attrs) > 0 {
		p.attrs, _ = applyListOverrides(defaultAttrs, config.Attrs)
	}

	return p
}

// Process processes the given image data.
// It always returns a value that the caller owns independently (either a new
// ImageData or a Ref() of the original), so the caller can Close() it without
// worrying about whether it is the same object as the input.
func (p *Processor) Process(o *options.Options, data imagedata.ImageData) (imagedata.ImageData, error) {
	if data.Format() != imagetype.SVG {
		return data.Ref(), nil
	}

	processed, err := p.sanitize(data)
	if err != nil {
		return nil, err
	}

	return processed, nil
}

// sanitize sanitizes the SVG data.
// It strips <script> and unsafe attributes (on* events).
// When no transformation is needed, it returns a Ref() of the input so the
// caller always gets independent ownership.
func (p *Processor) sanitize(data imagedata.ImageData) (imagedata.ImageData, error) {
	if !p.config.Sanitize {
		return data.Ref(), nil
	}

	doc, err := xmlparser.NewDocument(data.Reader())
	if err != nil {
		return nil, newSanitizeError(err)
	}

	// Expand DOCTYPE-declared entities so that the checks below see the
	// same values a consumer expanding them would.
	if err := doc.ReplaceEntities(); err != nil {
		return nil, newSanitizeError(err)
	}

	// Entities are already expanded, so the DOCTYPE is no longer needed.
	// Strip it to prevent various attacks like "Billion laughs".
	doc.FilterChildren(func(child xmlparser.Token) bool {
		_, isDoctype := child.(*xmlparser.Directive)
		return !isDoctype
	})

	// Sanitize the document's children
	doc.FilterChildNodes(func(el *xmlparser.Node) bool {
		return p.sanitizeElement(el)
	})

	buf, ok := pool.Get().(*bytes.Buffer)
	if !ok {
		return nil, newSanitizeError(errors.New("svg.Sanitize: failed to get buffer from pool"))
	}
	buf.Reset()

	cancel := func() {
		pool.Put(buf)
	}

	// Write the sanitized document to the buffer
	if err := doc.WriteTo(buf); err != nil {
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

// sanitizeElement sanitizes a single SVG element.
// It returns true if the element should be kept, false if it should be removed.
func (p *Processor) sanitizeElement(el *xmlparser.Node) bool {
	if el == nil {
		return false
	}

	tagName := el.Name.Local()

	// Strip <script>, <iframe>, <form>, <object> and <embed> in any namespace,
	// unless the user explicitly allowed them
	if inSet(deniedTags, tagName) && !inSet(p.tags, tagName) {
		return false
	}

	// Elements in the SVG namespace must be allowlisted.
	// Elements in other namespaces (editor data, XHTML inside <foreignObject>, ...)
	// are kept, only their attributes are filtered.
	isSVG := el.NSURI == "" || el.NSURI == svgNS
	if isSVG && !inSet(p.tags, tagName) {
		return false
	}

	// Strip SMIL animations that mutate link attributes
	if isUnsafeAnimation(el) {
		return false
	}

	rule := hrefRuleFor(tagName)

	// Keep only allowlisted attributes with safe values
	el.Attrs.Filter(func(attr *xmlparser.Attribute) bool {
		return p.isAllowedAttr(attr, isSVG) && isAllowedHref(attr, rule)
	})

	// Recurse into children
	el.FilterChildNodes(func(child *xmlparser.Node) bool {
		return p.sanitizeElement(child)
	})

	// Keep this element
	return true
}

// isAllowedAttr reports whether the attribute is allowed on an element.
// isSVG tells whether the element is in the SVG namespace.
func (p *Processor) isAllowedAttr(attr *xmlparser.Attribute, isSVG bool) bool {
	prefix, local := attr.Name.Split()

	var allowed bool

	switch {
	case prefix == "" && local == "xmlns", prefix == "xmlns":
		// Namespace declarations
		allowed = true
	case prefix == "":
		if isEventAttr(local) {
			break
		}

		if isSVG {
			allowed = inSet(p.attrs, local) || isDataOrAriaAttr(local)
		} else {
			allowed = true
		}
	default:
		switch attr.NSURI {
		case xlinkNS:
			allowed = inSet(xlinkAttrs, local)
		case xmlparser.NamespaceXML:
			allowed = inSet(xmlAttrs, local)
		default:
			allowed = !isEventAttr(local)
		}
	}

	return allowed && !hasUnsafeValue(attr.Value)
}

// isUnsafeAnimation reports whether el is a SMIL animation element that targets
// a link attribute (href, src). The animated value is not inspected.
func isUnsafeAnimation(el *xmlparser.Node) bool {
	tagName := el.Name.Local()
	if tagName != "animate" && tagName != "set" {
		return false
	}

	target, ok := el.Attrs.GetByString("attributeName")
	if !ok {
		return false
	}

	local := xmlparser.Name(strings.TrimSpace(target)).Local()
	_, unsafe := animationTargetAttrs[strings.ToLower(local)]

	return unsafe
}
