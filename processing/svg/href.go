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

// isSafeHref reports whether value is safe to keep as href/xlink:href: empty,
// a fragment/relative reference, or using one of allowedSchemes after normalization.
func isSafeHref(value string, allowedSchemes map[string]struct{}) bool {
	normalized := html.UnescapeString(strings.TrimSpace(value))

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
