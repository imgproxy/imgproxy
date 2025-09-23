package processing

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/auximageprovider"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/security"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type Context struct {
	// The context to check for timeouts and cancellations
	Ctx context.Context

	// Current image being processed
	Img *vips.Image

	// Processing options this pipeline runs with
	PO ProcessingOptions

	// Security options this pipeline runs with
	SecOps security.Options

	// Original image data
	ImgData imagedata.ImageData

	// The watermark image provider, if any watermarking is to be done.
	WatermarkProvider auximageprovider.Provider

	SrcWidth  int
	SrcHeight int
	Angle     int
	Flip      bool

	CropWidth   int
	CropHeight  int
	CropGravity GravityOptions

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

type Step func(c *Context) error
type Pipeline []Step

// Run runs the given pipeline with the given parameters
func (p Pipeline) Run(
	ctx context.Context,
	img *vips.Image,
	po ProcessingOptions,
	secops security.Options,
	imgdata imagedata.ImageData,
) error {
	pctx := p.newContext(ctx, img, po, secops, imgdata)
	pctx.CalcParams()

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

func (p Pipeline) newContext(
	ctx context.Context,
	img *vips.Image,
	po ProcessingOptions,
	secops security.Options,
	imgdata imagedata.ImageData,
) Context {
	pctx := Context{
		Ctx:     ctx,
		Img:     img,
		PO:      po,
		SecOps:  secops,
		ImgData: imgdata,

		WScale: 1.0,
		HScale: 1.0,

		DprScale:        1.0,
		VectorBaseScale: 1.0,

		CropGravity: po.CropGravity(),
	}

	// If crop gravity is not set, use the general gravity option
	if pctx.CropGravity.Type == options.GravityUnknown {
		pctx.CropGravity = po.Gravity()
	}

	return pctx
}
