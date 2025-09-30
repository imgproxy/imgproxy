package monitoring

// Metadata key names
const (
	MetaSourceImageURL    = "imgproxy.source_image_url"
	MetaSourceImageOrigin = "imgproxy.source_image_origin"
	MetaOptions           = "imgproxy.options"
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
