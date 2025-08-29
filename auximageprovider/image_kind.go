package auximageprovider

// ImageKind represents the kind of image stored in the provider.
type ImageKind string

const (
	// ImageKindUnknown represents an unknown image kind.
	ImageKindUnknown ImageKind = "auxiliary image"
	// ImageKindWatermark represents a watermark image.
	ImageKindWatermark ImageKind = "watermark"
	// ImageKindFallback represents a fallback image.
	ImageKindFallback ImageKind = "fallback image"
)
