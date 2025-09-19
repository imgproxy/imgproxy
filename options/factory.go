package options

// Presets is a map of preset names to their corresponding urlOptions
type Presets = map[string]urlOptions

// Factory creates Options instances
type Factory struct {
	config  *Config // Factory configuration
	presets Presets // Parsed presets
}

// NewFactory creates new Factory instance
func NewFactory(config *Config) (*Factory, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	f := &Factory{
		config:  config,
		presets: make(map[string]urlOptions),
	}

	if err := f.parsePresets(); err != nil {
		return nil, err
	}

	if err := f.validatePresets(); err != nil {
		return nil, err
	}

	return f, nil
}

func (f *Factory) IsSecurityOptionsAllowed() error {
	if f.config.AllowSecurityOptions {
		return nil
	}

	return newSecurityOptionsError()
}
