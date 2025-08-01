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

func (pctx *pipelineContext) calcScale(width, height int, po *options.ProcessingOptions) {
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

	pctx.dprScale = po.Dpr

	if !po.Enlarge && !pctx.imgtype.IsVector() {
		minShrink := math.Min(wshrink, hshrink)
		if minShrink < 1 {
			wshrink /= minShrink
			hshrink /= minShrink

			if !po.Extend.Enabled {
				pctx.dprScale /= minShrink
			}
		}

		// The minimum of wshrink and hshrink is the maximum dprScale value
		// that can be used without enlarging the image.
		pctx.dprScale = math.Min(pctx.dprScale, math.Min(wshrink, hshrink))
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

	wshrink /= pctx.dprScale
	hshrink /= pctx.dprScale

	if wshrink > srcW {
		wshrink = srcW
	}

	if hshrink > srcH {
		hshrink = srcH
	}

	pctx.wscale = 1.0 / wshrink
	pctx.hscale = 1.0 / hshrink
}

func (pctx *pipelineContext) calcSizes(widthToScale, heightToScale int, po *options.ProcessingOptions) {
	pctx.targetWidth = imath.Scale(po.Width, pctx.dprScale*po.ZoomWidth)
	pctx.targetHeight = imath.Scale(po.Height, pctx.dprScale*po.ZoomHeight)

	pctx.scaledWidth = imath.Scale(widthToScale, pctx.wscale)
	pctx.scaledHeight = imath.Scale(heightToScale, pctx.hscale)

	if po.ResizingType == options.ResizeFillDown && !po.Enlarge {
		diffW := float64(pctx.targetWidth) / float64(pctx.scaledWidth)
		diffH := float64(pctx.targetHeight) / float64(pctx.scaledHeight)

		switch {
		case diffW > diffH && diffW > 1.0:
			pctx.resultCropHeight = imath.Scale(pctx.scaledWidth, float64(pctx.targetHeight)/float64(pctx.targetWidth))
			pctx.resultCropWidth = pctx.scaledWidth

		case diffH > diffW && diffH > 1.0:
			pctx.resultCropWidth = imath.Scale(pctx.scaledHeight, float64(pctx.targetWidth)/float64(pctx.targetHeight))
			pctx.resultCropHeight = pctx.scaledHeight

		default:
			pctx.resultCropWidth = pctx.targetWidth
			pctx.resultCropHeight = pctx.targetHeight
		}
	} else {
		pctx.resultCropWidth = pctx.targetWidth
		pctx.resultCropHeight = pctx.targetHeight
	}

	if po.ExtendAspectRatio.Enabled && pctx.targetWidth > 0 && pctx.targetHeight > 0 {
		outWidth := imath.MinNonZero(pctx.scaledWidth, pctx.resultCropWidth)
		outHeight := imath.MinNonZero(pctx.scaledHeight, pctx.resultCropHeight)

		diffW := float64(pctx.targetWidth) / float64(outWidth)
		diffH := float64(pctx.targetHeight) / float64(outHeight)

		switch {
		case diffH > diffW:
			pctx.extendAspectRatioHeight = imath.Scale(outWidth, float64(pctx.targetHeight)/float64(pctx.targetWidth))
			pctx.extendAspectRatioWidth = outWidth

		case diffW > diffH:
			pctx.extendAspectRatioWidth = imath.Scale(outHeight, float64(pctx.targetWidth)/float64(pctx.targetHeight))
			pctx.extendAspectRatioHeight = outHeight
		}
	}
}

func (pctx *pipelineContext) limitScale(widthToScale, heightToScale int, po *options.ProcessingOptions) {
	maxresultDim := po.SecurityOptions.MaxResultDimension

	if maxresultDim <= 0 {
		return
	}

	outWidth := imath.MinNonZero(pctx.scaledWidth, pctx.resultCropWidth)
	outHeight := imath.MinNonZero(pctx.scaledHeight, pctx.resultCropHeight)

	if po.Extend.Enabled {
		outWidth = imath.Max(outWidth, pctx.targetWidth)
		outHeight = imath.Max(outHeight, pctx.targetHeight)
	} else if po.ExtendAspectRatio.Enabled {
		outWidth = imath.Max(outWidth, pctx.extendAspectRatioWidth)
		outHeight = imath.Max(outHeight, pctx.extendAspectRatioHeight)
	}

	if po.Padding.Enabled {
		outWidth += imath.ScaleToEven(po.Padding.Left, pctx.dprScale) + imath.ScaleToEven(po.Padding.Right, pctx.dprScale)
		outHeight += imath.ScaleToEven(po.Padding.Top, pctx.dprScale) + imath.ScaleToEven(po.Padding.Bottom, pctx.dprScale)
	}

	if maxresultDim > 0 && (outWidth > maxresultDim || outHeight > maxresultDim) {
		downScale := float64(maxresultDim) / float64(imath.Max(outWidth, outHeight))

		pctx.wscale *= downScale
		pctx.hscale *= downScale

		// Prevent scaling below 1px
		if minWScale := 1.0 / float64(widthToScale); pctx.wscale < minWScale {
			pctx.wscale = minWScale
		}
		if minHScale := 1.0 / float64(heightToScale); pctx.hscale < minHScale {
			pctx.hscale = minHScale
		}

		pctx.dprScale *= downScale

		// Recalculate the sizes after changing the scales
		pctx.calcSizes(widthToScale, heightToScale, po)
	}
}

func prepare(pctx *pipelineContext, img *vips.Image, po *options.ProcessingOptions, imgdata imagedata.ImageData) error {
	pctx.imgtype = imagetype.Unknown
	if imgdata != nil {
		pctx.imgtype = imgdata.Format()
	}

	pctx.srcWidth, pctx.srcHeight, pctx.angle, pctx.flip = extractMeta(img, po.Rotate, po.AutoRotate)

	pctx.cropWidth = calcCropSize(pctx.srcWidth, po.Crop.Width)
	pctx.cropHeight = calcCropSize(pctx.srcHeight, po.Crop.Height)

	widthToScale := imath.MinNonZero(pctx.cropWidth, pctx.srcWidth)
	heightToScale := imath.MinNonZero(pctx.cropHeight, pctx.srcHeight)

	pctx.calcScale(widthToScale, heightToScale, po)

	// The size of a vector image is not checked during download, yet it can be very large.
	// So we should scale it down to the maximum allowed resolution
	if !pctx.trimmed && imgdata != nil && imgdata.Format().IsVector() && !po.Enlarge {
		resolution := imath.Round((float64(img.Width()*img.Height()) * pctx.wscale * pctx.hscale))
		if resolution > po.SecurityOptions.MaxSrcResolution {
			scale := math.Sqrt(float64(po.SecurityOptions.MaxSrcResolution) / float64(resolution))
			pctx.wscale *= scale
			pctx.hscale *= scale
		}
	}

	pctx.calcSizes(widthToScale, heightToScale, po)

	pctx.limitScale(widthToScale, heightToScale, po)

	return nil
}
