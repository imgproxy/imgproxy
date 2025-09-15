package options

import (
	"maps"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// Presets is a map of preset names to their corresponding urlOptions
type Presets = map[string]urlOptions

// Factory creates ProcessingOptions instances
type Factory struct {
	config   *Config           // Factory configuration
	security *security.Checker // Security checker for generating security options
	presets  Presets           // Parsed presets
}

// NewFactory creates new Factory instance
func NewFactory(config *Config, security *security.Checker) (*Factory, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	f := &Factory{
		config:   config,
		security: security,
		presets:  make(map[string]urlOptions),
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
	po := ProcessingOptions{
		ResizingType:      ResizeFit,
		Width:             0,
		Height:            0,
		ZoomWidth:         1,
		ZoomHeight:        1,
		Gravity:           GravityOptions{Type: GravityCenter},
		Enlarge:           false,
		Extend:            ExtendOptions{Enabled: false, Gravity: GravityOptions{Type: GravityCenter}},
		ExtendAspectRatio: ExtendOptions{Enabled: false, Gravity: GravityOptions{Type: GravityCenter}},
		Padding:           PaddingOptions{Enabled: false},
		Trim:              TrimOptions{Enabled: false, Threshold: 10, Smart: true},
		Rotate:            0,
		Quality:           0,
		MaxBytes:          0,
		Format:            imagetype.Unknown,
		Background:        vips.Color{R: 255, G: 255, B: 255},
		Blur:              0,
		Sharpen:           0,
		Dpr:               1,
		Watermark:         WatermarkOptions{Opacity: 1, Position: GravityOptions{Type: GravityCenter}},
		StripMetadata:     f.config.StripMetadata,
		KeepCopyright:     f.config.KeepCopyright,
		StripColorProfile: f.config.StripColorProfile,
		AutoRotate:        f.config.AutoRotate,
		EnforceThumbnail:  f.config.EnforceThumbnail,
		ReturnAttachment:  f.config.ReturnAttachment,

		SkipProcessingFormats: append([]imagetype.Type(nil), f.config.SkipProcessingFormats...),
		UsedPresets:           make([]string, 0, len(f.config.Presets)),

		SecurityOptions: f.security.NewOptions(),

		// Basically, we need this to update ETag when `IMGPROXY_QUALITY` is changed
		defaultQuality: f.config.Quality,
	}

	po.FormatQuality = make(map[imagetype.Type]int, len(f.config.FormatQuality))
	maps.Copy(po.FormatQuality, f.config.FormatQuality)

	return &po
}
