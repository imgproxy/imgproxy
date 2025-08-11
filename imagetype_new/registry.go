package imagetype_new

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

// Registry holds the type registry
type Registry struct {
	types []*TypeDesc
}

// globalRegistry is the default registry instance
var globalRegistry = &Registry{}

// RegisterType registers a new image type in the global registry.
// It panics if the type already exists (i.e., if a TypeDesc is already registered for this Type).
func RegisterType(desc *TypeDesc) Type {
	return globalRegistry.RegisterType(desc)
}

// GetType returns the TypeDesc for the given Type.
// Returns nil if the type is not registered.
func GetType(t Type) *TypeDesc {
	return globalRegistry.GetType(t)
}

// RegisterType registers a new image type in this registry.
// It panics if the type already exists (i.e., if a TypeDesc is already registered for this Type).
func (r *Registry) RegisterType(desc *TypeDesc) Type {
	r.types = append(r.types, desc)
	return Type(len(r.types) - 1) // -1 is unknown
}

// GetType returns the TypeDesc for the given Type.
// Returns nil if the type is not registered.
func (r *Registry) GetType(t Type) *TypeDesc {
	if int(t) >= len(r.types) {
		return nil
	}
	return r.types[t]
}
