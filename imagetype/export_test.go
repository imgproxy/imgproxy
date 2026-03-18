package imagetype

func NewRegistry() *registry {
	return newRegistry()
}

func (r *registry) Detectors() []detector {
	return r.detectors
}
