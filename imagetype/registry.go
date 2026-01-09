package imagetype

import (
	"errors"
	"io"
	"slices"

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
	SupportsHDR           bool
}

// DetectFunc is a function that detects the image type from byte data
type DetectFunc func(r bufreader.ReadPeeker, ct, ext string) (Type, error)

// detector is a struct that holds a detection function and its priority
type detector struct {
	priority int        // priority of the detector, lower is better
	fn       DetectFunc // function that detects the image type
}

// Registry holds the type registry
type registry struct {
	detectors   []detector
	types       []*TypeDesc
	typesByName map[string]Type // maps type string to Type
}

// globalRegistry is the default registry instance
var globalRegistry = NewRegistry()

// NewRegistry creates a new image type registry.
func NewRegistry() *registry {
	return &registry{
		types:       nil,
		typesByName: make(map[string]Type),
	}
}

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

// GetTypeByName returns all registered image types.
func GetTypeByName(name string) (Type, bool) {
	return globalRegistry.GetTypeByName(name)
}

// RegisterType registers a new image type in this registry.
// It panics if the type already exists (i.e., if a TypeDesc is already registered for this Type).
func (r *registry) RegisterType(desc *TypeDesc) Type {
	r.types = append(r.types, desc)
	typ := Type(len(r.types)) // 0 is unknown
	r.typesByName[desc.String] = typ

	// NOTE: this is a special case for JPEG. The problem is that JPEG is using
	// several alternative extensions and processing_options.go is using extension to
	// find a type by key. There might be not the only case (e.g. ".tif/.tiff").
	// We need to handle this case in a more generic way.
	if desc.String == "jpeg" {
		// JPEG is a special case, we need to alias it
		r.typesByName["jpg"] = typ
	}

	return typ
}

// GetTypeDesc returns the TypeDesc for the given Type.
// Returns nil if the type is not registered.
func (r *registry) GetTypeDesc(t Type) *TypeDesc {
	if t <= 0 { // This would be "default" type
		return nil
	}

	if int(t-1) >= len(r.types) {
		return nil
	}

	return r.types[t-1]
}

// GetTypeByName returns Type by it's name
func (r *registry) GetTypeByName(name string) (Type, bool) {
	typ, ok := r.typesByName[name]
	return typ, ok
}

// RegisterDetector registers a custom detector function
// Detectors are tried in the order they were registered
func RegisterDetector(priority int, fn DetectFunc) {
	globalRegistry.RegisterDetector(priority, fn)
}

// RegisterMagicBytes registers magic bytes for a specific image type
// Magic byte detectors are always tried before custom detectors
func RegisterMagicBytes(typ Type, signature ...[]byte) {
	globalRegistry.RegisterMagicBytes(typ, signature...)
}

// Detect attempts to detect the image type from a reader.
// It first tries magic byte detection, then custom detectors in registration order
func Detect(r io.Reader, ct, ext string) (Type, error) {
	return globalRegistry.Detect(r, ct, ext)
}

// RegisterDetector registers a custom detector function on this registry instance
func (r *registry) RegisterDetector(priority int, fn DetectFunc) {
	r.detectors = append(r.detectors, detector{
		priority: priority,
		fn:       fn,
	})
	// Sort detectors by priority.
	// We don't expect a huge number of detectors, and detectors should be registered at startup,
	// so sorting them at each registration is okay.
	slices.SortStableFunc(r.detectors, func(a, b detector) int {
		return a.priority - b.priority
	})
}

// RegisterMagicBytes registers magic bytes for a specific image type on this registry instance
func (r *registry) RegisterMagicBytes(typ Type, signature ...[]byte) {
	r.RegisterDetector(-1, func(r bufreader.ReadPeeker, _, _ string) (Type, error) {
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
func (r *registry) Detect(re io.Reader, ct, ext string) (Type, error) {
	br := bufreader.New(io.LimitReader(re, maxDetectionLimit))

	for _, d := range r.detectors {
		br.Rewind()
		typ, err := d.fn(br, ct, ext)
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			return Unknown, newTypeDetectionError(err)
		}
		if err == nil && typ != Unknown {
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
