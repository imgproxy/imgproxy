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

	JPEG = RegisterType(&TypeDesc{
		String:                "jpeg",
		Ext:                   ".jpg",
		Mime:                  "image/jpeg",
		IsVector:              false,
		SupportsAlpha:         false,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	JXL = RegisterType(&TypeDesc{
		String:                "jxl",
		Ext:                   ".jxl",
		Mime:                  "image/jxl",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: true,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	PNG = RegisterType(&TypeDesc{
		String:                "png",
		Ext:                   ".png",
		Mime:                  "image/png",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       false,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	WEBP = RegisterType(&TypeDesc{
		String:                "webp",
		Ext:                   ".webp",
		Mime:                  "image/webp",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: true,
		SupportsAnimationSave: true,
		SupportsThumbnail:     false,
	})

	GIF = RegisterType(&TypeDesc{
		String:                "gif",
		Ext:                   ".gif",
		Mime:                  "image/gif",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       false,
		SupportsAnimationLoad: true,
		SupportsAnimationSave: true,
		SupportsThumbnail:     false,
	})

	ICO = RegisterType(&TypeDesc{
		String:                "ico",
		Ext:                   ".ico",
		Mime:                  "image/x-icon",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       false,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	SVG = RegisterType(&TypeDesc{
		String:                "svg",
		Ext:                   ".svg",
		Mime:                  "image/svg+xml",
		IsVector:              true,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       false,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	HEIC = RegisterType(&TypeDesc{
		String:                "heic",
		Ext:                   ".heic",
		Mime:                  "image/heif",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     true,
	})

	AVIF = RegisterType(&TypeDesc{
		String:                "avif",
		Ext:                   ".avif",
		Mime:                  "image/avif",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: true,
		SupportsQuality:       true,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     true,
	})

	BMP = RegisterType(&TypeDesc{
		String:                "bmp",
		Ext:                   ".bmp",
		Mime:                  "image/bmp",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       false,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})

	TIFF = RegisterType(&TypeDesc{
		String:                "tiff",
		Ext:                   ".tiff",
		Mime:                  "image/tiff",
		IsVector:              false,
		SupportsAlpha:         true,
		SupportsColourProfile: false,
		SupportsQuality:       true,
		SupportsAnimationLoad: false,
		SupportsAnimationSave: false,
		SupportsThumbnail:     false,
	})
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
