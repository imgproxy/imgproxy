package vips

/*
#include "vips.h"
*/
import "C"
import (
	"fmt"
	"os"

	globalConfig "github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ensure"
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

	c.JpegProgressive = globalConfig.JpegProgressive

	c.PngInterlaced = globalConfig.PngInterlaced
	c.PngQuantize = globalConfig.PngQuantize
	c.PngQuantizationColors = globalConfig.PngQuantizationColors

	if pr, ok := WebpPresets[globalConfig.WebpPreset]; ok {
		c.WebpPreset = pr
	} else {
		return nil, fmt.Errorf("invalid WebP preset: %s", globalConfig.WebpPreset)
	}

	c.AvifSpeed = globalConfig.AvifSpeed
	c.WebpEffort = globalConfig.WebpEffort
	c.JxlEffort = globalConfig.JxlEffort

	c.PngUnlimited = globalConfig.PngUnlimited
	c.SvgUnlimited = globalConfig.SvgUnlimited

	c.LeakCheck = len(os.Getenv("IMGPROXY_VIPS_LEAK_CHECK")) > 0
	c.CacheTrace = len(os.Getenv("IMGPROXY_VIPS_CACHE_TRACE")) > 0

	return c, nil
}

func (c *Config) Validate() error {
	if c.PngQuantizationColors < 2 || c.PngQuantizationColors > 256 {
		return fmt.Errorf(
			"IMGPROXY_PNG_QUANTIZATION_COLORS must be between 2 and 256, got %d",
			c.PngQuantizationColors,
		)
	}

	if c.WebpPreset < C.VIPS_FOREIGN_WEBP_PRESET_DEFAULT || c.WebpPreset >= C.VIPS_FOREIGN_WEBP_PRESET_LAST {
		return fmt.Errorf("invalid IMGPROXY_WEBP_PRESET: %d", c.WebpPreset)
	}

	if c.AvifSpeed < 0 || c.AvifSpeed > 9 {
		return fmt.Errorf("IMGPROXY_AVIF_SPEED must be between 0 and 9, got %d", c.AvifSpeed)
	}

	if c.JxlEffort < 1 || c.JxlEffort > 9 {
		return fmt.Errorf("IMGPROXY_JXL_EFFORT must be between 1 and 9, got %d", c.JxlEffort)
	}

	if c.WebpEffort < 1 || c.WebpEffort > 6 {
		return fmt.Errorf("IMGPROXY_WEBP_EFFORT must be between 1 and 6, got %d", c.WebpEffort)
	}

	return nil
}
