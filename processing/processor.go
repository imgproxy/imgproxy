package processing

import (
	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/processing/svg"
	"github.com/imgproxy/imgproxy/v3/security"
)

// Processor is responsible for processing images according to the given configuration.
type Processor struct {
	config            *Config
	securityChecker   *security.Checker
	watermarkProvider auximageprovider.Provider
	svg               *svg.Processor
}

// New creates a new Processor instance with the given configuration and watermark provider
func New(
	config *Config,
	securityChecker *security.Checker,
	watermark auximageprovider.Provider,
) (*Processor, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Processor{
		config:            config,
		securityChecker:   securityChecker,
		watermarkProvider: watermark,
		svg:               svg.New(&config.Svg),
	}, nil
}
