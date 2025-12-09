package processing

import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/processing/svg"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var (
	IMGPROXY_PREFERRED_FORMATS       = env.Describe("IMGPROXY_PREFERRED_FORMATS", "jpeg|png|gif|webp|avif|jxl|tiff|svg")
	IMGPROXY_SKIP_PROCESSING_FORMATS = env.Describe("IMGPROXY_SKIP_PROCESSING_FORMATS", "jpeg|png|gif|webp|avif|jxl|tiff|svg") //nolint:lll
	IMGPROXY_WATERMARK_OPACITY       = env.Describe("IMGPROXY_WATERMARK_OPACITY", "number between 0..1")
	IMGPROXY_DISABLE_SHRINK_ON_LOAD  = env.Describe("IMGPROXY_DISABLE_SHRINK_ON_LOAD", "boolean")
	IMGPROXY_USE_LINEAR_COLORSPACE   = env.Describe("IMGPROXY_USE_LINEAR_COLORSPACE", "boolean")
	IMGPROXY_ALWAYS_RASTERIZE_SVG    = env.Describe("IMGPROXY_ALWAYS_RASTERIZE_SVG", "boolean")
	IMGPROXY_QUALITY                 = env.Describe("IMGPROXY_QUALITY", "number between 0..100")
	IMGPROXY_FORMAT_QUALITY          = env.Describe("IMGPROXY_FORMAT_QUALITY", "comma-separated list of format=quality pairs where quality is between 0..100") //nolint:lll
	IMGPROXY_STRIP_METADATA          = env.Describe("IMGPROXY_STRIP_METADATA", "boolean")
	IMGPROXY_KEEP_COPYRIGHT          = env.Describe("IMGPROXY_KEEP_COPYRIGHT", "boolean")
	IMGPROXY_STRIP_COLOR_PROFILE     = env.Describe("IMGPROXY_STRIP_COLOR_PROFILE", "boolean")
	IMGPROXY_AUTO_ROTATE             = env.Describe("IMGPROXY_AUTO_ROTATE", "boolean")
	IMGPROXY_ENFORCE_THUMBNAIL       = env.Describe("IMGPROXY_ENFORCE_THUMBNAIL", "boolean")
)

// Config holds pipeline-related configuration.
type Config struct {
	PreferredFormats      []imagetype.Type
	SkipProcessingFormats []imagetype.Type
	WatermarkOpacity      float64
	DisableShrinkOnLoad   bool
	UseLinearColorspace   bool
	AlwaysRasterizeSvg    bool
	Quality               int
	FormatQuality         map[imagetype.Type]int
	StripMetadata         bool
	KeepCopyright         bool
	StripColorProfile     bool
	AutoRotate            bool
	EnforceThumbnail      bool

	Svg svg.Config
}

// NewDefaultConfig creates a new Config instance with the given parameters.
func NewDefaultConfig() Config {
	return Config{
		WatermarkOpacity: 1,
		PreferredFormats: []imagetype.Type{
			imagetype.JPEG,
			imagetype.PNG,
			imagetype.GIF,
		},
		Quality: 80,
		FormatQuality: map[imagetype.Type]int{
			imagetype.WEBP: 79,
			imagetype.AVIF: 63,
			imagetype.JXL:  77,
		},
		StripMetadata:     true,
		KeepCopyright:     true,
		StripColorProfile: true,
		AutoRotate:        true,
		EnforceThumbnail:  false,

		Svg: svg.NewDefaultConfig(),
	}
}

// LoadConfigFromEnv creates a new Config instance with the given parameters.
func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	_, svgErr := svg.LoadConfigFromEnv(&c.Svg)

	err := errors.Join(
		svgErr,
		env.Float(&c.WatermarkOpacity, IMGPROXY_WATERMARK_OPACITY),
		env.Bool(&c.DisableShrinkOnLoad, IMGPROXY_DISABLE_SHRINK_ON_LOAD),
		env.Bool(&c.UseLinearColorspace, IMGPROXY_USE_LINEAR_COLORSPACE),
		env.Bool(&c.AlwaysRasterizeSvg, IMGPROXY_ALWAYS_RASTERIZE_SVG),
		env.Int(&c.Quality, IMGPROXY_QUALITY),
		env.ImageTypesQuality(c.FormatQuality, IMGPROXY_FORMAT_QUALITY),
		env.Bool(&c.StripMetadata, IMGPROXY_STRIP_METADATA),
		env.Bool(&c.KeepCopyright, IMGPROXY_KEEP_COPYRIGHT),
		env.Bool(&c.StripColorProfile, IMGPROXY_STRIP_COLOR_PROFILE),
		env.Bool(&c.AutoRotate, IMGPROXY_AUTO_ROTATE),
		env.Bool(&c.EnforceThumbnail, IMGPROXY_ENFORCE_THUMBNAIL),

		env.ImageTypes(&c.PreferredFormats, IMGPROXY_PREFERRED_FORMATS),
		env.ImageTypes(&c.SkipProcessingFormats, IMGPROXY_SKIP_PROCESSING_FORMATS),
	)

	return c, err
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.WatermarkOpacity <= 0 || c.WatermarkOpacity > 1 {
		return IMGPROXY_WATERMARK_OPACITY.Errorf("must be between 0 and 1")
	}

	if c.Quality <= 0 || c.Quality > 100 {
		return IMGPROXY_QUALITY.Errorf("must be between 0 and 100")
	}

	filtered := c.PreferredFormats[:0]

	for _, t := range c.PreferredFormats {
		if !vips.SupportsSave(t) {
			IMGPROXY_PREFERRED_FORMATS.Warn("can't be a preferred format as it's saving is not supported", "format", t)
		} else {
			filtered = append(filtered, t)
		}
	}

	if len(filtered) == 0 {
		return IMGPROXY_PREFERRED_FORMATS.Errorf("no supported preferred formats specified")
	}

	c.PreferredFormats = filtered

	return nil
}
