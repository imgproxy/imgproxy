package security

// Security represents the security package instance
type Security struct {
	config *Config
}

// New creates a new Security instance
func New(config *Config) (*Security, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Security{
		config: config,
	}, nil
}

// NewOptions creates a new security.Options instance
func (s *Security) NewOptions() Options {
	return Options{
		MaxSrcResolution:            s.config.Options.MaxSrcResolution,
		MaxSrcFileSize:              s.config.Options.MaxSrcFileSize,
		MaxAnimationFrames:          s.config.Options.MaxAnimationFrames,
		MaxAnimationFrameResolution: s.config.Options.MaxAnimationFrameResolution,
		MaxResultDimension:          s.config.Options.MaxResultDimension,
	}
}
