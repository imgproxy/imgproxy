package monitoring

import (
	"context"
	"net/url"
	"strings"

	"github.com/imgproxy/imgproxy/v4/options"
	"github.com/imgproxy/imgproxy/v4/server/meta"
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

// contextKeys maps server/meta context keys to their monitoring metadata key equivalents.
var contextKeys = map[string]string{
	meta.KeyImageURL:          MetaSourceImageURL,
	meta.KeySourceImageOrigin: MetaSourceImageOrigin,
	meta.KeyOptions:           MetaOptions,
}

// NewMetaFromContext builds a Meta from the values server/meta has stored in ctx,
// translating each to its monitoring key name. Keys with no monitoring equivalent
// are dropped. Returns nil if ctx carries no *meta.Meta.
func NewMetaFromContext(ctx context.Context) Meta {
	cm := meta.FromContext(ctx)
	if cm == nil {
		return nil
	}

	return Meta(cm.Map(func(key string, value any) (string, any) {
		mk, ok := contextKeys[key]
		if !ok {
			return "", nil
		}

		// meta.KeyOptions holds *options.Options; monitoring wants its flat map form.
		if key == meta.KeyOptions {
			if o, ok := value.(*options.Options); ok {
				value = o.Map()
			}
		}

		return mk, value
	}))
}

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
