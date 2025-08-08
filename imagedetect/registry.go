package imagedetect

import (
	"sync"

	"github.com/imgproxy/imgproxy/v3/imagetype_new"
)

// DetectFunc is a function that detects the image type from byte data
type DetectFunc func(b []byte) (imagetype_new.Type, error)

// MagicBytes represents a magic byte signature for image type detection
// Signature can contain '?' characters which match any byte
type MagicBytes struct {
	Signature []byte
	Type      imagetype_new.Type
}

// Detector represents a registered detector function with its byte requirements
type Detector struct {
	Func        DetectFunc
	BytesNeeded int
}

// Registry manages the registration and execution of image type detectors
type Registry struct {
	mu         sync.RWMutex
	detectors  []Detector
	magicBytes []MagicBytes
}

// Global registry instance
var registry = &Registry{}

// RegisterDetector registers a custom detector function
// Detectors are tried in the order they were registered
func RegisterDetector(detector DetectFunc, bytesNeeded int) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.detectors = append(registry.detectors, Detector{
		Func:        detector,
		BytesNeeded: bytesNeeded,
	})
}

// RegisterMagicBytes registers magic bytes for a specific image type
// Magic byte detectors are always tried before custom detectors
func RegisterMagicBytes(signature []byte, typ imagetype_new.Type) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.magicBytes = append(registry.magicBytes, MagicBytes{
		Signature: signature,
		Type:      typ,
	})
}

// RegisterDetector registers a custom detector function on this registry instance
func (r *Registry) RegisterDetector(detector DetectFunc, bytesNeeded int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.detectors = append(r.detectors, Detector{
		Func:        detector,
		BytesNeeded: bytesNeeded,
	})
}

// RegisterMagicBytes registers magic bytes for a specific image type on this registry instance
func (r *Registry) RegisterMagicBytes(signature []byte, typ imagetype_new.Type) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.magicBytes = append(r.magicBytes, MagicBytes{
		Signature: signature,
		Type:      typ,
	})
}
