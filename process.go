package main

import (
	"math"

	"github.com/h2non/bimg"
)

type processingOptions struct {
	resize  string
	width   int
	height  int
	gravity bimg.Gravity
	enlarge bool
	format  bimg.ImageType
}

var imageTypes = map[string]bimg.ImageType{
	"jpeg": bimg.JPEG,
	"jpg":  bimg.JPEG,
	"png":  bimg.PNG,
}

var gravityTypes = map[string]bimg.Gravity{
	"ce": bimg.GravityCentre,
	"no": bimg.GravityNorth,
	"ea": bimg.GravityEast,
	"so": bimg.GravitySouth,
	"we": bimg.GravityWest,
	"sm": bimg.GravitySmart,
}

func round(f float64) int {
	return int(f + .5)
}
func calcSize(size bimg.ImageSize, po processingOptions) (int, int) {
	if (po.width == size.Width && po.height == size.Height) || (po.resize != "fill" && po.resize != "fit") {
		return po.width, po.height
	}

	fsw, fsh, fow, foh := float64(size.Width), float64(size.Height), float64(po.width), float64(po.height)

	wr := fow / fsw
	hr := foh / fsh

	var rate float64
	if po.resize == "fit" {
		rate = math.Min(wr, hr)
	} else {
		rate = math.Max(wr, hr)
	}

	return round(math.Min(fsw*rate, fow)), round(math.Min(fsh*rate, foh))
}

func processImage(p []byte, po processingOptions) ([]byte, error) {
	var err error

	img := bimg.NewImage(p)

	size, err := img.Size()
	if err != nil {
		return nil, err
	}

	// Default options
	opts := bimg.Options{
		Interpolator: bimg.Bicubic,
		Quality:      conf.Quality,
		Compression:  conf.Compression,
		Gravity:      po.gravity,
		Enlarge:      po.enlarge,
		Type:         po.format,
	}

	opts.Width, opts.Height = calcSize(size, po)

	switch po.resize {
	case "fit":
		opts.Embed = true
	case "fill":
		opts.Embed = true
		opts.Crop = true
	case "crop":
		opts.Crop = true
	default:
		opts.Force = true
	}

	return img.Process(opts)
}
