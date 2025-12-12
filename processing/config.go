package processing

import (
	"errors"
	"maps"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/processing/svg"
	"github.com/imgproxy/imgproxy/v3/vips"
)

var (
	IMGPROXY_PREFERRED_FORMATS       = env.ImageTypes("IMGPROXY_PREFERRED_FORMATS")
	IMGPROXY_SKIP_PROCESSING_FORMATS = env.ImageTypes("IMGPROXY_SKIP_PROCESSING_FORMATS")
	IMGPROXY_WATERMARK_OPACITY       = env.Float("IMGPROXY_WATERMARK_OPACITY")
	IMGPROXY_DISABLE_SHRINK_ON_LOAD  = env.Bool("IMGPROXY_DISABLE_SHRINK_ON_LOAD")
	IMGPROXY_USE_LINEAR_COLORSPACE   = env.Bool("IMGPROXY_USE_LINEAR_COLORSPACE")
	IMGPROXY_ALWAYS_RASTERIZE_SVG    = env.Bool("IMGPROXY_ALWAYS_RASTERIZE_SVG")
	IMGPROXY_QUALITY                 = env.Int("IMGPROXY_QUALITY")
	IMGPROXY_FORMAT_QUALITY          = env.ImageTypesQuality("IMGPROXY_FORMAT_QUALITY")
	IMGPROXY_STRIP_METADATA          = env.Bool("IMGPROXY_STRIP_METADATA")
	IMGPROXY_KEEP_COPYRIGHT          = env.Bool("IMGPROXY_KEEP_COPYRIGHT")
	IMGPROXY_STRIP_COLOR_PROFILE     = env.Bool("IMGPROXY_STRIP_COLOR_PROFILE")
	IMGPROXY_AUTO_ROTATE             = env.Bool("IMGPROXY_AUTO_ROTATE")
	IMGPROXY_ENFORCE_THUMBNAIL       = env.Bool("IMGPROXY_ENFORCE_THUMBNAIL")
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

	var fq map[imagetype.Type]int

	err := errors.Join(
		svgErr,
		IMGPROXY_WATERMARK_OPACITY.Parse(&c.WatermarkOpacity),
		IMGPROXY_DISABLE_SHRINK_ON_LOAD.Parse(&c.DisableShrinkOnLoad),
		IMGPROXY_USE_LINEAR_COLORSPACE.Parse(&c.UseLinearColorspace),
		IMGPROXY_ALWAYS_RASTERIZE_SVG.Parse(&c.AlwaysRasterizeSvg),
		IMGPROXY_QUALITY.Parse(&c.Quality),
		IMGPROXY_FORMAT_QUALITY.Parse(&fq),
		IMGPROXY_STRIP_METADATA.Parse(&c.StripMetadata),
		IMGPROXY_KEEP_COPYRIGHT.Parse(&c.KeepCopyright),
		IMGPROXY_STRIP_COLOR_PROFILE.Parse(&c.StripColorProfile),
		IMGPROXY_AUTO_ROTATE.Parse(&c.AutoRotate),
		IMGPROXY_ENFORCE_THUMBNAIL.Parse(&c.EnforceThumbnail),

		IMGPROXY_PREFERRED_FORMATS.Parse(&c.PreferredFormats),
		IMGPROXY_SKIP_PROCESSING_FORMATS.Parse(&c.SkipProcessingFormats),
	)

	maps.Copy(c.FormatQuality, fq)

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
