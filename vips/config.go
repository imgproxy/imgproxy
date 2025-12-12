package vips

/*
#include "vips.h"
*/
import "C"
import (
	"errors"

	"github.com/imgproxy/imgproxy/v3/ensure"
	"github.com/imgproxy/imgproxy/v3/env"
)

var (
	IMGPROXY_JPEG_PROGRESSIVE        = env.Bool("IMGPROXY_JPEG_PROGRESSIVE")
	IMGPROXY_PNG_INTERLACED          = env.Bool("IMGPROXY_PNG_INTERLACED")
	IMGPROXY_PNG_QUANTIZE            = env.Bool("IMGPROXY_PNG_QUANTIZE")
	IMGPROXY_PNG_QUANTIZATION_COLORS = env.Int("IMGPROXY_PNG_QUANTIZATION_COLORS")
	IMGPROXY_WEBP_PRESET             = env.Enum("IMGPROXY_WEBP_PRESET", WebpPresets)
	IMGPROXY_AVIF_SPEED              = env.Int("IMGPROXY_AVIF_SPEED")
	IMGPROXY_WEBP_EFFORT             = env.Int("IMGPROXY_WEBP_EFFORT")
	IMGPROXY_JXL_EFFORT              = env.Int("IMGPROXY_JXL_EFFORT")
	IMGPROXY_PNG_UNLIMITED           = env.Bool("IMGPROXY_PNG_UNLIMITED")
	IMGPROXY_SVG_UNLIMITED           = env.Bool("IMGPROXY_SVG_UNLIMITED")
	IMGPROXY_VIPS_LEAK_CHECK         = env.Bool("IMGPROXY_VIPS_LEAK_CHECK")
	IMGPROXY_VIPS_CACHE_TRACE        = env.Bool("IMGPROXY_VIPS_CACHE_TRACE")
)

type Config struct {
	// Whether to save JPEG as progressive
	JpegProgressive bool

	// Whether to save PNG as interlaced
	PngInterlaced bool
	// Whether to save PNG with adaptive palette
	PngQuantize bool
	// Number of colors for adaptive palette
	PngQuantizationColors int

	// WebP preset to use when saving WebP images
	WebpPreset WebpPreset

	// AVIF saving speed
	AvifSpeed int
	// WebP saving effort
	WebpEffort int
	// JPEG XL saving effort
	JxlEffort int

	// Whether to not apply any limits when loading PNG
	PngUnlimited bool
	// Whether to not apply any limits when loading JPEG
	SvgUnlimited bool

	// Whether to enable libvips memory leak check
	LeakCheck bool
	// Whether to enable libvips operation cache tracing
	CacheTrace bool
}

func NewDefaultConfig() Config {
	return Config{
		JpegProgressive: false,

		PngInterlaced:         false,
		PngQuantize:           false,
		PngQuantizationColors: 256,

		WebpPreset: C.VIPS_FOREIGN_WEBP_PRESET_DEFAULT,

		AvifSpeed:  8,
		WebpEffort: 4,
		JxlEffort:  4,

		PngUnlimited: false,
		SvgUnlimited: false,

		LeakCheck:  false,
		CacheTrace: false,
	}
}

func LoadConfigFromEnv(c *Config) (*Config, error) {
	c = ensure.Ensure(c, NewDefaultConfig)

	err := errors.Join(
		IMGPROXY_JPEG_PROGRESSIVE.Parse(&c.JpegProgressive),
		IMGPROXY_PNG_INTERLACED.Parse(&c.PngInterlaced),
		IMGPROXY_PNG_QUANTIZE.Parse(&c.PngQuantize),
		IMGPROXY_PNG_QUANTIZATION_COLORS.Parse(&c.PngQuantizationColors),
		IMGPROXY_AVIF_SPEED.Parse(&c.AvifSpeed),
		IMGPROXY_WEBP_EFFORT.Parse(&c.WebpEffort),
		IMGPROXY_JXL_EFFORT.Parse(&c.JxlEffort),
		IMGPROXY_PNG_UNLIMITED.Parse(&c.PngUnlimited),
		IMGPROXY_SVG_UNLIMITED.Parse(&c.SvgUnlimited),

		IMGPROXY_WEBP_PRESET.Parse(&c.WebpPreset),
		IMGPROXY_VIPS_LEAK_CHECK.Parse(&c.LeakCheck),
		IMGPROXY_VIPS_CACHE_TRACE.Parse(&c.CacheTrace),
	)

	return c, err
}

func (c *Config) Validate() error {
	if c.PngQuantizationColors < 2 || c.PngQuantizationColors > 256 {
		return IMGPROXY_PNG_QUANTIZATION_COLORS.ErrorRange()
	}

	if c.WebpPreset < C.VIPS_FOREIGN_WEBP_PRESET_DEFAULT || c.WebpPreset >= C.VIPS_FOREIGN_WEBP_PRESET_LAST {
		return IMGPROXY_WEBP_PRESET.ErrorRange()
	}

	if c.AvifSpeed < 0 || c.AvifSpeed > 9 {
		return IMGPROXY_AVIF_SPEED.ErrorRange()
	}

	if c.JxlEffort < 1 || c.JxlEffort > 9 {
		return IMGPROXY_JXL_EFFORT.ErrorRange()
	}

	if c.WebpEffort < 1 || c.WebpEffort > 6 {
		return IMGPROXY_WEBP_EFFORT.ErrorRange()
	}

	return nil
}
