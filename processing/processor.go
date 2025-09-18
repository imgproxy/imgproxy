package processing

import (
	"github.com/imgproxy/imgproxy/v3/auximageprovider"
)

// Processor is responsible for processing images according to the given configuration.
type Processor struct {
	config            *Config
	watermarkProvider auximageprovider.Provider
}

// New creates a new Processor instance with the given configuration and watermark provider
func New(config *Config, watermark auximageprovider.Provider) (*Processor, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Processor{
		config:            config,
		watermarkProvider: watermark,
	}, nil
}
