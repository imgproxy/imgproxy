package svg

import "strings"

// defaultAttrs lists unprefixed attributes allowed on SVG-namespace elements.
// Names are lowercase: matching is case-insensitive. The list is global,
// not per element, like in DOMPurify.
//
// The base is DOMPurify's `svg` list (src/attrs.ts). The tail adds attributes
// real-world files use and DOMPurify lacks (SMIL, lighting, SVG fonts, ...).
var defaultAttrs = []string{
	// DOMPurify svg
	"accent-height", "accumulate", "additive", "alignment-baseline",
	"amplitude", "ascent", "attributename", "attributetype", "azimuth",
	"basefrequency", "baseline-shift", "begin", "bias", "by", "class", "clip",
	"clippathunits", "clip-path", "clip-rule", "color", "color-interpolation",
	"color-interpolation-filters", "color-profile", "color-rendering", "cx",
	"cy", "d", "dx", "dy", "diffuseconstant", "direction", "display",
	"divisor", "dominant-baseline", "dur", "edgemode", "elevation", "end",
	"exponent", "fill", "fill-opacity", "fill-rule", "filter", "filterunits",
	"flood-color", "flood-opacity", "font-family", "font-size",
	"font-size-adjust", "font-stretch", "font-style", "font-variant",
	"font-weight", "fx", "fy", "g1", "g2", "glyph-name", "glyphref",
	"gradientunits", "gradienttransform", "height", "href", "id",
	"image-rendering", "in", "in2", "intercept", "k", "k1", "k2", "k3", "k4",
	"kerning", "keypoints", "keysplines", "keytimes", "lang", "lengthadjust",
	"letter-spacing", "kernelmatrix", "kernelunitlength", "lighting-color",
	"local", "marker-end", "marker-mid", "marker-start", "markerheight",
	"markerunits", "markerwidth", "maskcontentunits", "maskunits", "max",
	"mask", "mask-type", "media", "method", "mode", "min", "name",
	"numoctaves", "offset", "operator", "opacity", "order", "orient",
	"orientation", "origin", "overflow", "paint-order", "path", "pathlength",
	"patterncontentunits", "patterntransform", "patternunits",
	"pointer-events", "points", "preservealpha", "preserveaspectratio",
	"primitiveunits", "r", "rx", "ry", "radius", "refx", "refy", "repeatcount",
	"repeatdur", "restart", "result", "rotate", "scale", "seed",
	"shape-rendering", "slope", "specularconstant", "specularexponent",
	"spreadmethod", "startoffset", "stddeviation", "stitchtiles", "stop-color",
	"stop-opacity", "stroke-dasharray", "stroke-dashoffset", "stroke-linecap",
	"stroke-linejoin", "stroke-miterlimit", "stroke-opacity", "stroke",
	"stroke-width", "style", "surfacescale", "systemlanguage", "tabindex",
	"tablevalues", "targetx", "targety", "transform", "transform-origin",
	"text-anchor", "text-decoration", "text-orientation", "text-rendering",
	"textlength", "type", "u1", "u2", "unicode", "values", "vector-effect",
	"viewbox", "visibility", "version", "vert-adv-y", "vert-origin-x",
	"vert-origin-y", "width", "word-spacing", "wrap", "writing-mode",
	"xchannelselector", "ychannelselector", "x", "x1", "x2", "y", "y1", "y2",
	"z", "zoomandpan",

	// Not in DOMPurify
	"from", "to", "calcmode", "target", "rel", "role",
	"requiredfeatures", "requiredextensions", "baseprofile",
	"mix-blend-mode", "isolation", "enable-background", "white-space",
	"unicode-bidi", "fr", "contentstyletype", "transform-box",
	"pointsatx", "pointsaty", "pointsatz", "limitingconeangle",
	"glyph-orientation-horizontal", "glyph-orientation-vertical",
	"horiz-adv-x", "horiz-origin-x", "units-per-em", "unicode-range",
	"panose-1", "x-height", "cap-height", "bbox", "alphabetic", "hanging",
	"ideographic", "mathematical", "underline-position", "underline-thickness",
	"strikethrough-position", "strikethrough-thickness", "overline-position",
	"overline-thickness", "stemh", "stemv", "widths", "arabic-form",
	"rendering-intent", "viewtarget", "filterres", "cursor", "focusable",
	"font-variant-caps", "font-variant-east-asian", "font-variant-ligatures",
	"font-variant-numeric", "font-kerning", "text-overflow",
}

// defaultAttrSet is defaultAttrs as a set.
var defaultAttrSet = newSet(defaultAttrs)

// xlinkAttrs and xmlAttrs list allowed attributes in the xlink and xml
// namespaces, whatever prefix they are bound to.
var (
	xlinkAttrs = newSet([]string{"href", "title"})
	xmlAttrs   = newSet([]string{"space", "id", "lang"})
)

// unsafeEndTags are the elements whose closing tag in an attribute value can
// break out of the surrounding markup when the document is re-parsed.
var unsafeEndTags = strings.Fields(
	"style script title xmp textarea noscript iframe noembed noframes",
)

// hasUnsafeValue reports whether an attribute value can break out of the
// surrounding markup when the document is re-parsed: it closes a comment or
// CDATA section ("-->", "--!>", "]>") or contains a closing raw-text tag
// ("</script", "</style", ...). Taken from DOMPurify's SAFE_FOR_XML check.
//
// For example, `x--><script>alert(1)</script>` or `</style><script>...` as an
// attribute value would terminate the surrounding comment or raw-text element
// and start live markup.
//
// Most values (paths, colors, base64 data) contain neither '>' nor '<', so
// they cost two IndexByte calls.
func hasUnsafeValue(value string) bool {
	for offset := 0; offset < len(value); {
		i := strings.IndexByte(value[offset:], '>')
		if i < 0 {
			break
		}

		end := offset + i
		before := value[:end]

		if strings.HasSuffix(before, "]") ||
			strings.HasSuffix(before, "--") ||
			strings.HasSuffix(before, "--!") {
			return true
		}

		offset = end + 1
	}

	for offset := 0; offset < len(value); {
		i := strings.Index(value[offset:], "</")
		if i < 0 {
			break
		}

		start := offset + i + 2

		for _, tag := range unsafeEndTags {
			if hasPrefixFold(value[start:], tag) {
				return true
			}
		}

		offset = start
	}

	return false
}

// isEventAttr reports whether the (local) attribute name is an event handler
// (onclick, onload, ...). Event handlers run script, so they are dropped
// regardless of the attribute allowlist; matching on the "on" prefix also
// covers handlers that SVG or browsers may add later.
func isEventAttr(name string) bool {
	return hasPrefixFold(name, "on")
}

// isDataOrAriaAttr reports whether the attribute is data-* or aria-*.
// These families are open-ended (names are chosen by the author) and can't be
// listed in the allowlist, but they are inert, so they are allowed by prefix.
func isDataOrAriaAttr(name string) bool {
	return hasPrefixFold(name, "data-") || hasPrefixFold(name, "aria-")
}

// hasPrefixFold reports whether s starts with the lowercase ASCII prefix,
// ignoring ASCII case.
func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}

	for i := range len(prefix) {
		if lowerASCII(s[i]) != prefix[i] {
			return false
		}
	}

	return true
}

func lowerASCII(c byte) byte {
	if 'A' <= c && c <= 'Z' {
		return c + 'a' - 'A'
	}

	return c
}

func newSet(names []string) map[string]struct{} {
	set := make(map[string]struct{}, len(names))

	for _, name := range names {
		set[name] = struct{}{}
	}

	return set
}

// inSet reports whether the name, lowercased, is in the set of lowercase names.
func inSet(set map[string]struct{}, name string) bool {
	_, ok := set[strings.ToLower(name)]

	return ok
}
