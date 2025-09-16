package options

import (
	"github.com/imgproxy/imgproxy/v3/security"
)

// Presets is a map of preset names to their corresponding urlOptions
type Presets = map[string]urlOptions

// Factory creates ProcessingOptions instances
type Factory struct {
	config    *Config            // Factory configuration
	security  *security.Checker  // Security checker for generating security options
	presets   Presets            // Parsed presets
	defaultPO *ProcessingOptions // Default processing options
}

// NewFactory creates new Factory instance
func NewFactory(config *Config, security *security.Checker) (*Factory, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	f := &Factory{
		config:    config,
		security:  security,
		presets:   make(map[string]urlOptions),
		defaultPO: newDefaultProcessingOptions(config, security),
	}

	if err := f.parsePresets(); err != nil {
		return nil, err
	}

	if err := f.validatePresets(); err != nil {
		return nil, err
	}

	return f, nil
}

// NewProcessingOptions creates new ProcessingOptions instance
func (f *Factory) NewProcessingOptions() *ProcessingOptions {
	return f.defaultPO.clone()
}
