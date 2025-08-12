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
var (
	// Unknown is a reserved type, it has index 0. We guarantee that index 0 won't be used
	// for any other type. This way, Unknown is a zero value for Type.
	Unknown Type = 0
)

// Mime returns the MIME type for the image type.
func (t Type) Mime() string {
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.Mime
	}

	return "application/octet-stream"
}

// String returns the string representation of the image type.
func (t Type) String() string {
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.String
	}

	return ""
}

// Ext returns the file extension for the image type.
func (t Type) Ext() string {
	desc := GetTypeDesc(t)
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
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.IsVector
	}
	return false
}

// SupportsAlpha checks if the image type supports alpha transparency.
func (t Type) SupportsAlpha() bool {
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.SupportsAlpha
	}
	return true
}

// SupportsAnimationLoad checks if the image type supports animation.
func (t Type) SupportsAnimationLoad() bool {
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.SupportsAnimationLoad
	}
	return false
}

// SupportsAnimationSave checks if the image type supports saving animations.
func (t Type) SupportsAnimationSave() bool {
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.SupportsAnimationSave
	}
	return false
}

// SupportsColourProfile checks if the image type supports metadata.
func (t Type) SupportsColourProfile() bool {
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.SupportsColourProfile
	}
	return false
}

// SupportsQuality checks if the image type supports quality adjustments.
func (t Type) SupportsQuality() bool {
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.SupportsQuality
	}
	return false
}

// SupportsThumbnail checks if the image type supports thumbnails.
func (t Type) SupportsThumbnail() bool {
	desc := GetTypeDesc(t)
	if desc != nil {
		return desc.SupportsThumbnail
	}
	return false
}
