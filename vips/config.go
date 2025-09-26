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
	IMGPROXY_JPEG_PROGRESSIVE        = env.Describe("IMGPROXY_JPEG_PROGRESSIVE", "boolean")
	IMGPROXY_PNG_INTERLACED          = env.Describe("IMGPROXY_PNG_INTERLACED", "boolean")
	IMGPROXY_PNG_QUANTIZE            = env.Describe("IMGPROXY_PNG_QUANTIZE", "boolean")
	IMGPROXY_PNG_QUANTIZATION_COLORS = env.Describe("IMGPROXY_PNG_QUANTIZATION_COLORS", "number between 2 and 256")
	IMGPROXY_WEBP_PRESET             = env.Describe("IMGPROXY_WEBP_PRESET", "default|picture|photo|drawing|icon|text")
	IMGPROXY_AVIF_SPEED              = env.Describe("IMGPROXY_AVIF_SPEED", "number between 0 and 9")
	IMGPROXY_WEBP_EFFORT             = env.Describe("IMGPROXY_WEBP_EFFORT", "number between 1 and 6")
	IMGPROXY_JXL_EFFORT              = env.Describe("IMGPROXY_JXL_EFFORT", "number between 1 and 9")
	IMGPROXY_PNG_UNLIMITED           = env.Describe("IMGPROXY_PNG_UNLIMITED", "boolean")
	IMGPROXY_SVG_UNLIMITED           = env.Describe("IMGPROXY_SVG_UNLIMITED", "boolean")
	IMGPROXY_VIPS_LEAK_CHECK         = env.Describe("IMGPROXY_VIPS_LEAK_CHECK", "boolean")
	IMGPROXY_VIPS_CACHE_TRACE        = env.Describe("IMGPROXY_VIPS_CACHE_TRACE", "boolean")
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

	var leakCheck, cacheTrace string

	// default preset so parsing below won't fail on empty value
	webpPreset := c.WebpPreset.String()

	err := errors.Join(
		env.Bool(&c.JpegProgressive, IMGPROXY_JPEG_PROGRESSIVE),
		env.Bool(&c.PngInterlaced, IMGPROXY_PNG_INTERLACED),
		env.Bool(&c.PngQuantize, IMGPROXY_PNG_QUANTIZE),
		env.Int(&c.PngQuantizationColors, IMGPROXY_PNG_QUANTIZATION_COLORS),
		env.Int(&c.AvifSpeed, IMGPROXY_AVIF_SPEED),
		env.Int(&c.WebpEffort, IMGPROXY_WEBP_EFFORT),
		env.Int(&c.JxlEffort, IMGPROXY_JXL_EFFORT),
		env.Bool(&c.PngUnlimited, IMGPROXY_PNG_UNLIMITED),
		env.Bool(&c.SvgUnlimited, IMGPROXY_SVG_UNLIMITED),

		env.String(&webpPreset, IMGPROXY_WEBP_PRESET),
		env.String(&leakCheck, IMGPROXY_VIPS_LEAK_CHECK),
		env.String(&cacheTrace, IMGPROXY_VIPS_CACHE_TRACE),
	)
	if err != nil {
		return nil, err
	}

	if pr, ok := WebpPresets[webpPreset]; ok {
		c.WebpPreset = pr
	} else {
		return nil, IMGPROXY_WEBP_PRESET.Errorf("invalid WebP preset: %s", webpPreset)
	}

	c.LeakCheck = len(leakCheck) > 0
	c.CacheTrace = len(cacheTrace) > 0

	return c, nil
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
