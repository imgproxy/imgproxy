package imagetype_new

import (
	"fmt"
	"sync"
)

// TypeDesc is used to store metadata about an image type.
// It represents the minimal information needed to make imgproxy to
// work with the type.
type TypeDesc struct {
	String                string
	Ext                   string
	Mime                  string
	IsVector              bool
	SupportsAlpha         bool
	SupportsColourProfile bool
	SupportsQuality       bool
	SupportsAnimationLoad bool
	SupportsAnimationSave bool
	SupportsThumbnail     bool
}

// Registry holds the type registry and mutex for thread-safe operations
type Registry struct {
	types []*TypeDesc
	mu    sync.Mutex
}

// globalRegistry is the default registry instance
var globalRegistry = &Registry{}

// RegisterType registers a new image type in the global registry.
// It panics if the type already exists (i.e., if a TypeDesc is already registered for this Type).
func RegisterType(t Type, desc *TypeDesc) {
	err := globalRegistry.RegisterType(t, desc)
	if err != nil {
		panic(err)
	}
}

// GetType returns the TypeDesc for the given Type.
// Returns nil if the type is not registered.
func GetType(t Type) *TypeDesc {
	return globalRegistry.GetType(t)
}

// RegisterType registers a new image type in this registry.
// It panics if the type already exists (i.e., if a TypeDesc is already registered for this Type).
func (r *Registry) RegisterType(t Type, desc *TypeDesc) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Ensure the registry is large enough to hold this type
	for len(r.types) <= int(t) {
		r.types = append(r.types, nil)
	}

	// Check if type already exists
	if r.types[t] != nil {
		return fmt.Errorf("type %d is already registered", t)
	}

	// Register the type
	r.types[t] = desc

	return nil
}

// GetType returns the TypeDesc for the given Type.
// Returns nil if the type is not registered.
func (r *Registry) GetType(t Type) *TypeDesc {
	// No mutex needed for reading as types are only modified during startup
	if int(t) >= len(r.types) {
		return nil
	}
	return r.types[t]
}

// init registers all default image types
func init() {
	RegisterType(JPEG, &TypeDesc{
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
	RegisterType(JXL, &TypeDesc{
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
	RegisterType(PNG, &TypeDesc{
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
	RegisterType(WEBP, &TypeDesc{
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
	RegisterType(GIF, &TypeDesc{
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
	RegisterType(ICO, &TypeDesc{
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
	RegisterType(SVG, &TypeDesc{
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
	RegisterType(HEIC, &TypeDesc{
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
	RegisterType(AVIF, &TypeDesc{
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
	RegisterType(BMP, &TypeDesc{
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
	RegisterType(TIFF, &TypeDesc{
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
}
