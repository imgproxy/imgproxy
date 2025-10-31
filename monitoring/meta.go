package monitoring

import (
	"net/url"
	"strings"
)

// Metadata key names
const (
	MetaPrefix            = "imgproxy."
	MetaSourceImageURL    = MetaPrefix + "source_image_url"
	MetaSourceImageOrigin = MetaPrefix + "source_image_origin"
	MetaOptions           = MetaPrefix + "options"
)

// Meta represents a set of metadata key-value pairs.
type Meta map[string]any

// Filter creates a copy of Meta with only the specified keys.
func (m Meta) Filter(only ...string) Meta {
	filtered := make(Meta)
	for _, key := range only {
		if value, ok := m[key]; ok {
			filtered[key] = value
		}
	}
	return filtered
}

// MetaKey formats a metadata key with the standard prefix.
func MetaKey(key string) string {
	return MetaPrefix + strings.ToLower(strings.ReplaceAll(key, " ", "_"))
}

// MetaURLOrigin extracts the origin (scheme + host) from a URL for metadata purposes.
func MetaURLOrigin(fullURL string) string {
	if u, err := url.Parse(fullURL); err == nil {
		return u.Scheme + "://" + u.Host
	}

	return ""
}
