package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

func extractMeta(img *vips.Image, baseAngle int, useOrientation bool) (int, int, int, bool) {
	width := img.Width()
	height := img.Height()

	angle := 0
	flip := false

	if useOrientation {
		orientation := img.Orientation()

		if orientation == 3 || orientation == 4 {
			angle = 180
		}
		if orientation == 5 || orientation == 6 {
			angle = 90
		}
		if orientation == 7 || orientation == 8 {
			angle = 270
		}
		if orientation == 2 || orientation == 4 || orientation == 5 || orientation == 7 {
			flip = true
		}
	}

	if (angle+baseAngle)%180 != 0 {
		width, height = height, width
	}

	return width, height, angle, flip
}

func calcScale(width, height int, po *options.ProcessingOptions, imgtype imagetype.Type) (float64, float64) {
	var wshrink, hshrink float64

	srcW, srcH := float64(width), float64(height)
	dstW, dstH := float64(po.Width), float64(po.Height)

	if po.Width == 0 {
		dstW = srcW
	}

	if dstW == srcW {
		wshrink = 1
	} else {
		wshrink = srcW / dstW
	}

	if po.Height == 0 {
		dstH = srcH
	}

	if dstH == srcH {
		hshrink = 1
	} else {
		hshrink = srcH / dstH
	}

	wshrink /= po.Dpr
	hshrink /= po.Dpr

	if wshrink != 1 || hshrink != 1 {
		rt := po.ResizingType

		if rt == options.ResizeAuto {
			srcD := srcW - srcH
			dstD := dstW - dstH

			if (srcD >= 0 && dstD >= 0) || (srcD < 0 && dstD < 0) {
				rt = options.ResizeFill
			} else {
				rt = options.ResizeFit
			}
		}

		switch {
		case po.Width == 0 && rt != options.ResizeForce:
			wshrink = hshrink
		case po.Height == 0 && rt != options.ResizeForce:
			hshrink = wshrink
		case rt == options.ResizeFit:
			wshrink = math.Max(wshrink, hshrink)
			hshrink = wshrink
		case rt == options.ResizeFill || rt == options.ResizeFillDown:
			wshrink = math.Min(wshrink, hshrink)
			hshrink = wshrink
		}
	}

	wshrink /= po.ZoomWidth
	hshrink /= po.ZoomHeight

	if !po.Enlarge && imgtype != imagetype.SVG {
		if wshrink < 1 {
			hshrink /= wshrink
			wshrink = 1
		}
		if hshrink < 1 {
			wshrink /= hshrink
			hshrink = 1
		}
	}

	if po.MinWidth > 0 {
		if minShrink := srcW / float64(po.MinWidth); minShrink < wshrink {
			hshrink /= wshrink / minShrink
			wshrink = minShrink
		}
	}

	if po.MinHeight > 0 {
		if minShrink := srcH / float64(po.MinHeight); minShrink < hshrink {
			wshrink /= hshrink / minShrink
			hshrink = minShrink
		}
	}

	if wshrink > srcW {
		wshrink = srcW
	}

	if hshrink > srcH {
		hshrink = srcH
	}

	return 1.0 / wshrink, 1.0 / hshrink
}

func calcCropSize(orig int, crop float64) int {
	switch {
	case crop == 0.0:
		return 0
	case crop >= 1.0:
		return int(crop)
	default:
		return imath.Max(1, imath.Scale(orig, crop))
	}
}

func prepare(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata *imagedata.ImageData) error {
	pctx.imgtype = imagetype.Unknown
	if imgdata != nil {
		pctx.imgtype = imgdata.Type
	}

	pctx.srcWidth, pctx.srcHeight, pctx.angle, pctx.flip = extractMeta(img, po.Rotate, po.AutoRotate)

	pctx.cropWidth = calcCropSize(pctx.srcWidth, po.Crop.Width)
	pctx.cropHeight = calcCropSize(pctx.srcHeight, po.Crop.Height)

	widthToScale := imath.MinNonZero(pctx.cropWidth, pctx.srcWidth)
	heightToScale := imath.MinNonZero(pctx.cropHeight, pctx.srcHeight)

	pctx.wscale, pctx.hscale = calcScale(widthToScale, heightToScale, po, pctx.imgtype)

	return nil
}
