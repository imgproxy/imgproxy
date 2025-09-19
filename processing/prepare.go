package processing

import (
	"math"

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
		return max(1, imath.Scale(orig, crop))
	}
}

func (pctx *Context) calcScale(width, height int, po *options.ProcessingOptions) {
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

	pctx.DprScale = po.Dpr

	isVector := pctx.ImgData != nil && pctx.ImgData.Format().IsVector()

	if !po.Enlarge && !isVector {
		minShrink := math.Min(wshrink, hshrink)
		if minShrink < 1 {
			wshrink /= minShrink
			hshrink /= minShrink

			// If we reached this point, this means that we can't reach the target size
			// because the image is smaller than it, and the enlargement is disabled.
			// If the DprScale is less than 1, the image will be downscaled, moving
			// even further from the target size, so we need to compensate it.
			// The compensation may increase the DprScale too much, but this is okay,
			// because we'll handle this further in the code.
			//
			// If the Extend option is enabled, we want to keep the resulting image
			// composition the same regardless of the DPR, so we don't apply this compensation
			// in this case.
			if !po.Extend.Enabled {
				pctx.DprScale /= minShrink
			}
		}

		// The minimum of wshrink and hshrink is the maximum dprScale value
		// that can be used without enlarging the image.
		pctx.DprScale = math.Min(pctx.DprScale, math.Min(wshrink, hshrink))
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

	wshrink /= pctx.DprScale
	hshrink /= pctx.DprScale

	if wshrink > srcW {
		wshrink = srcW
	}

	if hshrink > srcH {
		hshrink = srcH
	}

	pctx.WScale = 1.0 / wshrink
	pctx.HScale = 1.0 / hshrink
}

func (pctx *Context) calcSizes(widthToScale, heightToScale int, po *options.ProcessingOptions) {
	pctx.TargetWidth = imath.Scale(po.Width, pctx.DprScale*po.ZoomWidth)
	pctx.TargetHeight = imath.Scale(po.Height, pctx.DprScale*po.ZoomHeight)

	pctx.ScaledWidth = imath.Scale(widthToScale, pctx.WScale)
	pctx.ScaledHeight = imath.Scale(heightToScale, pctx.HScale)

	if po.ResizingType == options.ResizeFillDown && !po.Enlarge {
		diffW := float64(pctx.TargetWidth) / float64(pctx.ScaledWidth)
		diffH := float64(pctx.TargetHeight) / float64(pctx.ScaledHeight)

		switch {
		case diffW > diffH && diffW > 1.0:
			pctx.ResultCropHeight = imath.Scale(pctx.ScaledWidth, float64(pctx.TargetHeight)/float64(pctx.TargetWidth))
			pctx.ResultCropWidth = pctx.ScaledWidth

		case diffH > diffW && diffH > 1.0:
			pctx.ResultCropWidth = imath.Scale(pctx.ScaledHeight, float64(pctx.TargetWidth)/float64(pctx.TargetHeight))
			pctx.ResultCropHeight = pctx.ScaledHeight

		default:
			pctx.ResultCropWidth = pctx.TargetWidth
			pctx.ResultCropHeight = pctx.TargetHeight
		}
	} else {
		pctx.ResultCropWidth = pctx.TargetWidth
		pctx.ResultCropHeight = pctx.TargetHeight
	}

	if po.ExtendAspectRatio.Enabled && pctx.TargetWidth > 0 && pctx.TargetHeight > 0 {
		outWidth := imath.MinNonZero(pctx.ScaledWidth, pctx.ResultCropWidth)
		outHeight := imath.MinNonZero(pctx.ScaledHeight, pctx.ResultCropHeight)

		diffW := float64(pctx.TargetWidth) / float64(outWidth)
		diffH := float64(pctx.TargetHeight) / float64(outHeight)

		switch {
		case diffH > diffW:
			pctx.ExtendAspectRatioHeight = imath.Scale(outWidth, float64(pctx.TargetHeight)/float64(pctx.TargetWidth))
			pctx.ExtendAspectRatioWidth = outWidth

		case diffW > diffH:
			pctx.ExtendAspectRatioWidth = imath.Scale(outHeight, float64(pctx.TargetWidth)/float64(pctx.TargetHeight))
			pctx.ExtendAspectRatioHeight = outHeight
		}
	}
}

func (pctx *Context) limitScale(widthToScale, heightToScale int, po *options.ProcessingOptions) {
	maxresultDim := po.SecurityOptions.MaxResultDimension

	if maxresultDim <= 0 {
		return
	}

	outWidth := imath.MinNonZero(pctx.ScaledWidth, pctx.ResultCropWidth)
	outHeight := imath.MinNonZero(pctx.ScaledHeight, pctx.ResultCropHeight)

	if po.Extend.Enabled {
		outWidth = max(outWidth, pctx.TargetWidth)
		outHeight = max(outHeight, pctx.TargetHeight)
	} else if po.ExtendAspectRatio.Enabled {
		outWidth = max(outWidth, pctx.ExtendAspectRatioWidth)
		outHeight = max(outHeight, pctx.ExtendAspectRatioHeight)
	}

	if po.Padding.Enabled {
		outWidth += imath.ScaleToEven(po.Padding.Left, pctx.DprScale) + imath.ScaleToEven(po.Padding.Right, pctx.DprScale)
		outHeight += imath.ScaleToEven(po.Padding.Top, pctx.DprScale) + imath.ScaleToEven(po.Padding.Bottom, pctx.DprScale)
	}

	if maxresultDim > 0 && (outWidth > maxresultDim || outHeight > maxresultDim) {
		downScale := float64(maxresultDim) / float64(max(outWidth, outHeight))

		pctx.WScale *= downScale
		pctx.HScale *= downScale

		// Prevent scaling below 1px
		if minWScale := 1.0 / float64(widthToScale); pctx.WScale < minWScale {
			pctx.WScale = minWScale
		}
		if minHScale := 1.0 / float64(heightToScale); pctx.HScale < minHScale {
			pctx.HScale = minHScale
		}

		pctx.DprScale *= downScale

		// Recalculate the sizes after changing the scales
		pctx.calcSizes(widthToScale, heightToScale, po)
	}
}

// prepare extracts image metadata and calculates scaling factors and target sizes.
// This can't be done in advance because some steps like trimming and rasterization could
// happen before this step.
func prepare(c *Context) error {
	c.SrcWidth, c.SrcHeight, c.Angle, c.Flip = extractMeta(c.Img, c.PO.Rotate, c.PO.AutoRotate)

	c.CropWidth = calcCropSize(c.SrcWidth, c.PO.Crop.Width)
	c.CropHeight = calcCropSize(c.SrcHeight, c.PO.Crop.Height)

	widthToScale := imath.MinNonZero(c.CropWidth, c.SrcWidth)
	heightToScale := imath.MinNonZero(c.CropHeight, c.SrcHeight)

	c.calcScale(widthToScale, heightToScale, c.PO)
	c.calcSizes(widthToScale, heightToScale, c.PO)
	c.limitScale(widthToScale, heightToScale, c.PO)

	return nil
}
