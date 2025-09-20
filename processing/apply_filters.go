package processing

import (
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
)

func applyFilters(c *Context) error {
	blur := options.GetFloat(c.PO, keys.Blur, 0.0)
	sharpen := options.GetFloat(c.PO, keys.Sharpen, 0.0)
	pixelate := options.GetInt(c.PO, keys.Pixelate, 1)

	if blur == 0 && sharpen == 0 && pixelate <= 1 {
		return nil
	}

	if err := c.Img.CopyMemory(); err != nil {
		return err
	}

	if err := c.Img.RgbColourspace(); err != nil {
		return err
	}

	if err := c.Img.ApplyFilters(blur, sharpen, pixelate); err != nil {
		return err
	}

	return c.Img.CopyMemory()
}
