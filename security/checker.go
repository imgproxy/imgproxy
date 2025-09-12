package security

// Checker represents the security package instance
type Checker struct {
	config *Config
}

// New creates a new Security instance
func New(config *Config) (*Checker, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Checker{
		config: config,
	}, nil
}

// NewOptions creates a new security.Options instance
func (s *Checker) NewOptions() Options {
	return s.config.DefaultOptions
}
