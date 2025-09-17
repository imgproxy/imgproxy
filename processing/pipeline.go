package processing

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type pipelineContext struct {
	ctx context.Context

	config *Config

	imgtype imagetype.Type

	// The watermark image provider, if any watermarking is to be done.
	watermarkProvider auximageprovider.Provider

	trimmed bool

	srcWidth  int
	srcHeight int
	angle     int
	flip      bool

	cropWidth   int
	cropHeight  int
	cropGravity options.GravityOptions

	wscale float64
	hscale float64

	dprScale float64

	// The base scale factor for vector images.
	// It is used to downscale the input vector image to the maximum allowed resolution
	vectorBaseScale float64

	// The width we aim to get.
	// Based on the requested width scaled according to processing options.
	// Can be 0 if width is not specified in the processing options.
	targetWidth int
	// The height we aim to get.
	// Based on the requested height scaled according to processing options.
	// Can be 0 if height is not specified in the processing options.
	targetHeight int

	// The width of the image after cropping, scaling and rotating
	scaledWidth int
	// The height of the image after cropping, scaling and rotating
	scaledHeight int

	// The width of the result crop according to the resizing type
	resultCropWidth int
	// The height of the result crop according to the resizing type
	resultCropHeight int

	// The width of the image extended to the requested aspect ratio.
	// Can be 0 if any of the dimensions is not specified in the processing options
	// or if the image already has the requested aspect ratio.
	extendAspectRatioWidth int
	// The width of the image extended to the requested aspect ratio.
	// Can be 0 if any of the dimensions is not specified in the processing options
	// or if the image already has the requested aspect ratio.
	extendAspectRatioHeight int
}

type pipelineStep func(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata imagedata.ImageData) error
type pipeline []pipelineStep

func (p pipeline) Run(
	ctx context.Context,
	img *vips.Image,
	po *options.ProcessingOptions,
	imgdata imagedata.ImageData,
	watermark auximageprovider.Provider,
	config *Config,
) error {
	pctx := pipelineContext{
		ctx:    ctx,
		config: config,

		wscale: 1.0,
		hscale: 1.0,

		dprScale:        1.0,
		vectorBaseScale: 1.0,

		cropGravity:       po.Crop.Gravity,
		watermarkProvider: watermark,
	}

	if pctx.cropGravity.Type == options.GravityUnknown {
		pctx.cropGravity = po.Gravity
	}

	for _, step := range p {
		if err := step(&pctx, img, po, imgdata); err != nil {
			return err
		}

		if err := server.CheckTimeout(ctx); err != nil {
			return err
		}
	}

	img.SetDouble("imgproxy-dpr-scale", pctx.dprScale)

	return nil
}
