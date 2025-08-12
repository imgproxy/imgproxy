package imagetype_new

import (
	"io"

	"github.com/imgproxy/imgproxy/v3/bufreader"
)

const (
	// maxDetectionLimit is maximum bytes detectors allowed to read from the source
	maxDetectionLimit = 32 * 1024
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

// DetectFunc is a function that detects the image type from byte data
type DetectFunc func(r bufreader.ReadPeeker) (Type, error)

// Registry holds the type registry
type Registry struct {
	detectors []DetectFunc
	types     []*TypeDesc
}

// globalRegistry is the default registry instance
var globalRegistry = &Registry{}

// RegisterType registers a new image type in the global registry.
// It panics if the type already exists (i.e., if a TypeDesc is already registered for this Type).
func RegisterType(desc *TypeDesc) Type {
	return globalRegistry.RegisterType(desc)
}

// GetTypeDesc returns the TypeDesc for the given Type.
// Returns nil if the type is not registered.
func GetTypeDesc(t Type) *TypeDesc {
	return globalRegistry.GetTypeDesc(t)
}

// RegisterType registers a new image type in this registry.
// It panics if the type already exists (i.e., if a TypeDesc is already registered for this Type).
func (r *Registry) RegisterType(desc *TypeDesc) Type {
	r.types = append(r.types, desc)
	return Type(len(r.types)) // 0 is unknown
}

// GetTypeDesc returns the TypeDesc for the given Type.
// Returns nil if the type is not registered.
func (r *Registry) GetTypeDesc(t Type) *TypeDesc {
	if t <= 0 { // This would be "default" type
		return nil
	}

	if int(t-1) >= len(r.types) {
		return nil
	}

	return r.types[t-1]
}

// RegisterDetector registers a custom detector function
// Detectors are tried in the order they were registered
func RegisterDetector(detector DetectFunc) {
	globalRegistry.RegisterDetector(detector)
}

// RegisterMagicBytes registers magic bytes for a specific image type
// Magic byte detectors are always tried before custom detectors
func RegisterMagicBytes(typ Type, signature ...[]byte) {
	globalRegistry.RegisterMagicBytes(typ, signature...)
}

// Detect attempts to detect the image type from a reader.
// It first tries magic byte detection, then custom detectors in registration order
func Detect(r io.Reader) (Type, error) {
	return globalRegistry.Detect(r)
}

// RegisterDetector registers a custom detector function on this registry instance
func (r *Registry) RegisterDetector(detector DetectFunc) {
	r.detectors = append(r.detectors, detector)
}

// RegisterMagicBytes registers magic bytes for a specific image type on this registry instance
func (r *Registry) RegisterMagicBytes(typ Type, signature ...[]byte) {
	r.detectors = append(r.detectors, func(r bufreader.ReadPeeker) (Type, error) {
		for _, sig := range signature {
			b, err := r.Peek(len(sig))
			if err != nil {
				return Unknown, err
			}

			if hasMagicBytes(b, sig) {
				return typ, nil
			}
		}

		return Unknown, nil
	})
}

// Detect runs image format detection
func (r *Registry) Detect(re io.Reader) (Type, error) {
	br := bufreader.New(io.LimitReader(re, maxDetectionLimit))

	for _, fn := range globalRegistry.detectors {
		br.Rewind()
		if typ, err := fn(br); err == nil && typ != Unknown {
			return typ, nil
		}
	}

	return Unknown, newUnknownFormatError()
}

// hasMagicBytes checks if the data matches a magic byte signature
// Supports '?' characters in signature which match any byte
func hasMagicBytes(data []byte, magic []byte) bool {
	if len(data) < len(magic) {
		return false
	}

	for i, c := range magic {
		if c != data[i] && c != '?' {
			return false
		}
	}
	return true
}
