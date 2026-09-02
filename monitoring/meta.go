package monitoring

import (
	"context"
	"net/url"
	"slices"
	"strings"

	"github.com/imgproxy/imgproxy/v4/options"
	"github.com/imgproxy/imgproxy/v4/server/meta"
)

// MetaPrefix is prepended to every monitoring metadata key.
const MetaPrefix = "imgproxy."

// Meta represents a set of metadata key-value pairs.
type Meta map[string]any

// NewMetaFromContext builds a Meta from the values server/meta has stored in ctx,
// translating each key to its monitoring key name via MetaKey. meta.KeyReqID has
// no monitoring equivalent and is always dropped. If keys is non-empty, only
// server/meta keys named in it are included. Returns nil if ctx carries no *meta.Meta.
func NewMetaFromContext(ctx context.Context, keys ...string) Meta {
	if meta.FromContext(ctx) == nil {
		return nil
	}

	return Meta(meta.Map(ctx, func(key string, value any) (string, any) {
		if len(keys) > 0 && !slices.Contains(keys, key) {
			return "", nil
		}

		// meta.KeyOptions holds *options.Options; monitoring wants its flat map form.
		if o, ok := value.(*options.Options); ok {
			value = o.Map()
		}

		return MetaKey(key), value
	}))
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
