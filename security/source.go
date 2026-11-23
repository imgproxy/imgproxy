package security

import (
	"strings"
)

// VerifySourceURL checks if the given URL is allowed to be used:
//  1. Checks that the path does not contain dot-only segments.
//  2. Checks that the URL is in the list of allowed sources, if any are configured.
//
// Returns an error if the URL is not allowed.
//
// Parameters:
//   - fullURL: the complete URL.
//   - urlPath: the path extracted from the URL. It should be properly percent-decoded.
func (s *Checker) VerifySourceURL(fullURL, urlPath string) error {
	if s.urlHasDotOnlySegment(urlPath) {
		return newSourceURLError(fullURL, "path contains a dot-only segment")
	}

	if !s.isURLAllowed(fullURL) {
		return newSourceURLError(fullURL, "not in allowed sources")
	}

	return nil
}

// isURLAllowed checks if the given URL is allowed based on the configured AllowedSources.
func (s *Checker) isURLAllowed(fullURL string) bool {
	if len(s.config.AllowedSources) == 0 {
		return true
	}

	for _, allowedSource := range s.config.AllowedSources {
		if allowedSource.MatchString(fullURL) {
			return true
		}
	}

	return false
}

// urlHasDotOnlySegment checks if the given URL path contains any dot-only segments (e.g., `/../`)
func (s *Checker) urlHasDotOnlySegment(urlPath string) bool {
	for seg := range strings.SplitSeq(urlPath, "/") {
		if len(seg) > 1 && len(strings.TrimLeft(seg, ".")) == 0 {
			return true
		}
	}

	return false
}
