package responsewriter

import "net/http"

// Factory is a struct that creates response writers.
type Factory struct {
	config *Config
}

func NewFactory(config *Config) (*Factory, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Factory{config}, nil
}

// NewWriter wraps [http.ResponseWriter] into [Writer].
func (f *Factory) NewWriter(rw http.ResponseWriter) *Writer {
	w := &Writer{
		config:        f.config,
		result:        make(http.Header),
		originHeaders: make(http.Header),
		maxAge:        -1,
	}

	w.SetHTTPResponseWriter(rw)

	return w
}
