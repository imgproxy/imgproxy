package processing

import (
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func trim(c *Context) error {
	if !options.Get(c.PO, keys.TrimEnabled, false) {
		return nil
	}

	// We need to import color profile before trim
	if err := colorspaceToProcessing(c); err != nil {
		return err
	}

	if err := c.Img.Trim(
		options.GetFloat(c.PO, keys.TrimThreshold, 10.0),
		options.Get(c.PO, keys.TrimSmart, true),
		options.Get(c.PO, keys.TrimColor, vips.Color{}),
		options.Get(c.PO, keys.TrimEqualHor, false),
		options.Get(c.PO, keys.TrimEqualVer, false),
	); err != nil {
		return err
	}
	if err := c.Img.CopyMemory(); err != nil {
		return err
	}

	c.ImgData = nil

	return nil
}
