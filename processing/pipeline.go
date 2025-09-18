package processing

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/processing/pipeline"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// NOTE: this will be called pipeline.Context in the separate package
type Context struct {
	Ctx context.Context

	// Global processing configuration which could be used by individual steps
	Config *pipeline.Config

	// VIPS image
	Img *vips.Image

	// Processing options this pipeline runs with
	PO *options.ProcessingOptions

	// Source image data
	ImgData imagedata.ImageData

	// The watermark image provider, if any watermarking is to be done.
	WatermarkProvider auximageprovider.Provider

	SrcWidth  int
	SrcHeight int
	Angle     int
	Flip      bool

	CropWidth   int
	CropHeight  int
	CropGravity options.GravityOptions

	WScale float64
	HScale float64

	DprScale float64

	// The base scale factor for vector images.
	// It is used to downscale the input vector image to the maximum allowed resolution
	VectorBaseScale float64

	// The width we aim to get.
	// Based on the requested width scaled according to processing options.
	// Can be 0 if width is not specified in the processing options.
	TargetWidth int
	// The height we aim to get.
	// Based on the requested height scaled according to processing options.
	// Can be 0 if height is not specified in the processing options.
	TargetHeight int

	// The width of the image after cropping, scaling and rotating
	ScaledWidth int
	// The height of the image after cropping, scaling and rotating
	ScaledHeight int

	// The width of the result crop according to the resizing type
	ResultCropWidth int
	// The height of the result crop according to the resizing type
	ResultCropHeight int

	// The width of the image extended to the requested aspect ratio.
	// Can be 0 if any of the dimensions is not specified in the processing options
	// or if the image already has the requested aspect ratio.
	ExtendAspectRatioWidth int
	// The width of the image extended to the requested aspect ratio.
	// Can be 0 if any of the dimensions is not specified in the processing options
	// or if the image already has the requested aspect ratio.
	ExtendAspectRatioHeight int
}

// NOTE: same, pipeline.Step, pipeline.Pipeline, pipeline.Runner
type Step func(ctx *Context) error
type Pipeline []Step

// Runner is responsible for running a processing pipeline
type Runner struct {
	config    *pipeline.Config
	watermark auximageprovider.Provider
}

// New creates a new Runner instance with the given configuration and watermark provider
func New(config *pipeline.Config, watermark auximageprovider.Provider) *Runner {
	return &Runner{
		config:    config,
		watermark: watermark,
	}
}

// Run runs the given pipeline with the given parameters
func (f *Runner) Run(
	p Pipeline,
	ctx context.Context,
	img *vips.Image,
	po *options.ProcessingOptions,
	imgdata imagedata.ImageData,
) error {
	pctx := f.newContext(ctx, img, po, imgdata)

	for _, step := range p {
		if err := step(&pctx); err != nil {
			return err
		}

		if err := server.CheckTimeout(ctx); err != nil {
			return err
		}
	}

	img.SetDouble("imgproxy-dpr-scale", pctx.DprScale)

	return nil
}

func (r *Runner) newContext(
	ctx context.Context,
	img *vips.Image,
	po *options.ProcessingOptions,
	imgdata imagedata.ImageData,
) Context {
	pctx := Context{
		Ctx:     ctx,
		Config:  r.config,
		Img:     img,
		PO:      po,
		ImgData: imgdata,

		WScale: 1.0,
		HScale: 1.0,

		DprScale:        1.0,
		VectorBaseScale: 1.0,

		CropGravity:       po.Crop.Gravity,
		WatermarkProvider: r.watermark,
	}

	if pctx.CropGravity.Type == options.GravityUnknown {
		pctx.CropGravity = po.Gravity
	}

	return pctx
}
