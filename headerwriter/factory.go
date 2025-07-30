package headerwriter

import "net/http"

// Factory is a struct that provides methods to create HeaderBuilder instances
type Factory struct {
	config *Config
}

// NewFactory creates a new factory instance with the provided configuration
func NewFactory(config *Config) *Factory {
	return &Factory{config: config}
}

// NewHeaderBuilder creates a new HeaderBuilder instance with the provided request headers
// NOTE: should URL be string here?
func (f *Factory) NewHeaderWriter(originalResponseHeaders http.Header, url string) *Writer {
	return newWriter(f.config, originalResponseHeaders, url)
}
