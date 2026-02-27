package conditionalheaders

import "net/http"

// Factory is responsible for creating Request instances from user requests.
type Factory struct {
	config *Config
}

// NewFactory creates a new Factory instance with the given configuration.
func NewFactory(c *Config) (*Factory, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &Factory{
		config: c,
	}, nil
}

// NewRequest creates a new Request instance from the given user request.
func (p *Factory) NewRequest(req *http.Request) *Request {
	return newFromRequest(p.config, req)
}
