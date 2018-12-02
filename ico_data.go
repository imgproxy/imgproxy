package main

import (
	"bytes"
	"image"
	"image/draw"

	_ "github.com/mat/besticon/ico"
)

func icoData(data []byte) (out []byte, width int, height int, err error) {
	var ico image.Image

	ico, _, err = image.Decode(bytes.NewReader(data))
	if err != nil {
		return
	}

	// Ensure that image is in RGBA format
	rgba := image.NewRGBA(ico.Bounds())
	draw.Draw(rgba, ico.Bounds(), ico, image.ZP, draw.Src)

	width = rgba.Bounds().Dx()
	height = rgba.Bounds().Dy()
	out = rgba.Pix

	return
}
