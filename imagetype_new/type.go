package imagetype_new

import (
	"fmt"
)

// Type represents an image type
type (
	// Type represents an image type.
	Type int
)

// Supported image types
const (
	Unknown Type = iota
	JPEG
	JXL
	PNG
	WEBP
	GIF
	ICO
	SVG
	HEIC
	AVIF
	BMP
	TIFF
)

// Mime returns the MIME type for the image type.
func (t Type) Mime() string {
	desc := GetType(t)
	if desc != nil {
		return desc.Mime
	}

	return "octet-stream"
}

// String returns the string representation of the image type.
func (t Type) String() string {
	desc := GetType(t)
	if desc != nil {
		return desc.String
	}

	return ""
}

// Ext returns the file extension for the image type.
func (t Type) Ext() string {
	desc := GetType(t)
	if desc != nil {
		return desc.Ext
	}

	return ""
}

// MarshalJSON implements the json.Marshaler interface for Type.
func (t Type) MarshalJSON() ([]byte, error) {
	s := t.String()
	if s == "" {
		return []byte("null"), nil
	}

	return fmt.Appendf(nil, "%q", s), nil
}

// IsVector checks if the image type is a vector format.
func (t Type) IsVector() bool {
	desc := GetType(t)
	if desc != nil {
		return desc.IsVector
	}
	return false
}

// SupportsAlpha checks if the image type supports alpha transparency.
func (t Type) SupportsAlpha() bool {
	desc := GetType(t)
	if desc != nil {
		return desc.SupportsAlpha
	}
	return false
}

// SupportsAnimationLoad checks if the image type supports animation.
func (t Type) SupportsAnimationLoad() bool {
	desc := GetType(t)
	if desc != nil {
		return desc.SupportsAnimationLoad
	}
	return false
}

// SupportsAnimationSave checks if the image type supports saving animations.
func (t Type) SupportsAnimationSave() bool {
	desc := GetType(t)
	if desc != nil {
		return desc.SupportsAnimationSave
	}
	return false
}

// SupportsMetadata checks if the image type supports metadata.
func (t Type) SupportsColourProfile() bool {
	desc := GetType(t)
	if desc != nil {
		return desc.SupportsColourProfile
	}
	return false
}

// SupportsQuality checks if the image type supports quality adjustments.
func (t Type) SupportsQuality() bool {
	desc := GetType(t)
	if desc != nil {
		return desc.SupportsQuality
	}
	return false
}

// SupportsThumbnail checks if the image type supports thumbnails.
func (t Type) SupportsThumbnail() bool {
	desc := GetType(t)
	if desc != nil {
		return desc.SupportsThumbnail
	}
	return false
}
