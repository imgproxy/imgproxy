package svg

import (
	"html"
	"regexp"
	"strings"
)

// hrefSchemes lists schemes allowed in href/xlink:href on generic elements.
var hrefSchemes = map[string]struct{}{
	"http":  {},
	"https": {},
}

// imageHrefSchemes is hrefSchemes plus "data", since <image> data: URIs are
// allowed.
var imageHrefSchemes = map[string]struct{}{
	"http":  {},
	"https": {},
	"data":  {},
}

// schemeRe matches a leading URI scheme, e.g. "https:" in "https://example.com".
var schemeRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9+.-]*:`)

// normalizeHref decodes entities and then strips what browsers strip before
// parsing a URL: tab, LF and CR anywhere, and leading/trailing C0 controls
// and spaces. Without this, "java&#9;script:" would look like a relative
// reference while browsers treat it as "javascript:".
func normalizeHref(value string) string {
	decoded := html.UnescapeString(value)

	decoded = strings.Map(func(r rune) rune {
		if r == '\t' || r == '\n' || r == '\r' {
			return -1
		}
		return r
	}, decoded)

	return strings.TrimFunc(decoded, func(r rune) bool {
		return r <= 0x20
	})
}

// isSafeHref reports whether value is safe to keep as href/xlink:href: empty,
// a fragment/relative reference, or using one of allowedSchemes after normalization.
func isSafeHref(value string, allowedSchemes map[string]struct{}) bool {
	normalized := normalizeHref(value)

	if len(normalized) == 0 || normalized[0] == '#' {
		return true
	}

	scheme := schemeRe.FindString(normalized)
	if len(scheme) == 0 {
		// No scheme found: relative or protocol-relative reference.
		return true
	}

	scheme = strings.ToLower(strings.TrimSuffix(scheme, ":"))
	_, allowed := allowedSchemes[scheme]

	return allowed
}
