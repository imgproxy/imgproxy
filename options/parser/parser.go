package optionsparser

// Presets is a map of preset names to their corresponding urlOptions
type Presets = map[string]urlOptions

// Parser creates Options instances
type Parser struct {
	config  *Config // Parser configuration
	presets Presets // Parsed presets
}

// New creates new Parser instance
func New(config *Config) (*Parser, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	p := &Parser{
		config:  config,
		presets: make(map[string]urlOptions),
	}

	if err := p.parsePresets(); err != nil {
		return nil, err
	}

	if err := p.validatePresets(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Parser) IsSecurityOptionsAllowed() error {
	if p.config.AllowSecurityOptions {
		return nil
	}

	return newSecurityOptionsError()
}
