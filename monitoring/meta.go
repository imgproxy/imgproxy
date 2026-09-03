package monitoring

import (
	"context"
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
// translating each key to its monitoring key name via MetaKey. If keys is non-empty, only
// server/meta keys named in it are included. Returns an empty, non-nil Meta if ctx
// carries no server/meta metadata.
func NewMetaFromContext(ctx context.Context, keys ...string) Meta {
	return Meta(meta.Map(ctx, func(key string, value any) (string, any) {
		if len(keys) > 0 && !slices.Contains(keys, key) {
			return "", nil
		}

		// monitoring wants options in its flat map form.
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
