package processing

import (
	"context"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

type pipelineContext struct {
	ctx context.Context

	imgtype imagetype.Type

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

	iccImported bool
}

type pipelineStep func(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error
type pipeline []pipelineStep

func (p pipeline) Run(ctx context.Context, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	pctx := pipelineContext{
		ctx: ctx,

		wscale: 1.0,
		hscale: 1.0,

		cropGravity: po.Crop.Gravity,
	}

	if pctx.cropGravity.Type == options.GravityUnknown {
		pctx.cropGravity = po.Gravity
	}

	for _, step := range p {
		if err := step(&pctx, img, po, imgdata); err != nil {
			return err
		}
	}

	return nil
}
