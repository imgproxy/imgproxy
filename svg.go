package main

import (
	"image"
	goColor "image/color"
	"io"
)

func init() {
	// Register fake svg decoder. Since we need this only for type detecting, we can
	// return fake image sizes
	image.RegisterFormat(
		"svg",
		"<?xml ",
		func(io.Reader) (image.Image, error) {
			return image.NewRGBA(image.Rect(0, 0, 1, 1)), nil
		},
		func(io.Reader) (image.Config, error) {
			return image.Config{ColorModel: goColor.RGBAModel, Width: 1, Height: 1}, nil
		},
	)
}
