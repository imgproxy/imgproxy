package imagedetect

import (
	"github.com/imgproxy/imgproxy/v3/bufreader"
	"github.com/imgproxy/imgproxy/v3/imagetype_new"
)

// DetectFunc is a function that detects the image type from byte data
type DetectFunc func(r bufreader.ReadPeeker) (imagetype_new.Type, error)

// Registry manages the registration and execution of image type detectors
type Registry struct {
	detectors []DetectFunc
}

// Global registry instance
var registry = &Registry{}

// RegisterDetector registers a custom detector function
// Detectors are tried in the order they were registered
func RegisterDetector(detector DetectFunc, bytesNeeded int) {
	registry.RegisterDetector(detector, bytesNeeded)
}

// RegisterMagicBytes registers magic bytes for a specific image type
// Magic byte detectors are always tried before custom detectors
func RegisterMagicBytes(signature []byte, typ imagetype_new.Type) {
	registry.RegisterMagicBytes(signature, typ)
}

// RegisterDetector registers a custom detector function on this registry instance
func (r *Registry) RegisterDetector(detector DetectFunc, bytesNeeded int) {
	r.detectors = append(r.detectors, detector)
}

// RegisterMagicBytes registers magic bytes for a specific image type on this registry instance
func (r *Registry) RegisterMagicBytes(signature []byte, typ imagetype_new.Type) {
	r.detectors = append(r.detectors, func(r bufreader.ReadPeeker) (imagetype_new.Type, error) {
		b, err := r.Peek(len(signature))
		if err != nil {
			return imagetype_new.Unknown, err
		}

		if hasMagicBytes(b, signature) {
			return typ, nil
		}

		return imagetype_new.Unknown, nil
	})
}
