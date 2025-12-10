package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// ExtractGeometry extracts image width, height, orientation angle and flip flag from the image metadata.
func ExtractGeometry(img *vips.Image, baseAngle int, autoRotate bool) (int, int, int, bool) {
	width := img.Width()
	height := img.Height()

	angle, flip := angleFlip(img, autoRotate)

	if (angle+baseAngle)%180 != 0 {
		width, height = height, width
	}

	return width, height, angle, flip
}

// angleFlip returns the orientation angle and flip flag based on the image metadata
// and po.AutoRotate flag.
func angleFlip(img *vips.Image, autoRotate bool) (int, bool) {
	if !autoRotate {
		return 0, false
	}

	angle := 0
	flip := false

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

	return angle, flip
}

// CalcCropSize calculates the crop size based on the original size and crop scale.
func CalcCropSize(orig int, crop float64) int {
	switch {
	case crop == 0.0:
		return 0
	case crop >= 1.0:
		return int(crop)
	default:
		return max(1, imath.Scale(orig, crop))
	}
}

func (c *Context) calcScale(width, height int, po ProcessingOptions) {
	wshrink, hshrink := 1.0, 1.0
	srcW, srcH := float64(width), float64(height)

	poWidth := po.Width()
	poHeight := po.Height()

	dstW := imath.NonZero(float64(poWidth), srcW)
	dstH := imath.NonZero(float64(poHeight), srcH)

	if dstW != srcW {
		wshrink = srcW / dstW
	}

	if dstH != srcH {
		hshrink = srcH / dstH
	}

	if wshrink != 1 || hshrink != 1 {
		rt := po.ResizingType()

		if rt == ResizeAuto {
			srcD := srcW - srcH
			dstD := dstW - dstH

			if (srcD >= 0 && dstD >= 0) || (srcD < 0 && dstD < 0) {
				rt = ResizeFill
			} else {
				rt = ResizeFit
			}
		}

		switch {
		case poWidth == 0 && rt != ResizeForce:
			wshrink = hshrink
		case poHeight == 0 && rt != ResizeForce:
			hshrink = wshrink
		case rt == ResizeFit:
			wshrink = math.Max(wshrink, hshrink)
			hshrink = wshrink
		case rt == ResizeFill || rt == ResizeFillDown:
			wshrink = math.Min(wshrink, hshrink)
			hshrink = wshrink
		}
	}

	wshrink /= po.ZoomWidth()
	hshrink /= po.ZoomHeight()

	c.DprScale = po.DPR()

	isVector := c.ImgData != nil && c.ImgData.Format().IsVector()

	if !po.Enlarge() && !isVector {
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
			if !po.ExtendEnabled() {
				c.DprScale /= minShrink
			}
		}

		// The minimum of wshrink and hshrink is the maximum dprScale value
		// that can be used without enlarging the image.
		c.DprScale = math.Min(c.DprScale, math.Min(wshrink, hshrink))
	}

	if minWidth := po.MinWidth(); minWidth > 0 {
		if minShrink := srcW / float64(minWidth); minShrink < wshrink {
			hshrink /= wshrink / minShrink
			wshrink = minShrink
		}
	}

	if minHeight := po.MinHeight(); minHeight > 0 {
		if minShrink := srcH / float64(minHeight); minShrink < hshrink {
			wshrink /= hshrink / minShrink
			hshrink = minShrink
		}
	}

	wshrink /= c.DprScale
	hshrink /= c.DprScale

	if wshrink > srcW {
		wshrink = srcW
	}

	if hshrink > srcH {
		hshrink = srcH
	}

	c.WScale = 1.0 / wshrink
	c.HScale = 1.0 / hshrink
}

func (c *Context) calcSizes(widthToScale, heightToScale int, po ProcessingOptions) {
	c.TargetWidth = imath.Scale(po.Width(), c.DprScale*po.ZoomWidth())
	c.TargetHeight = imath.Scale(po.Height(), c.DprScale*po.ZoomHeight())

	c.ScaledWidth = imath.Scale(widthToScale, c.WScale)
	c.ScaledHeight = imath.Scale(heightToScale, c.HScale)

	if po.ResizingType() == ResizeFillDown && !po.Enlarge() {
		diffW := float64(c.TargetWidth) / float64(c.ScaledWidth)
		diffH := float64(c.TargetHeight) / float64(c.ScaledHeight)

		switch {
		case diffW > diffH && diffW > 1.0:
			c.ResultCropHeight = imath.Scale(c.ScaledWidth, float64(c.TargetHeight)/float64(c.TargetWidth))
			c.ResultCropWidth = c.ScaledWidth

		case diffH > diffW && diffH > 1.0:
			c.ResultCropWidth = imath.Scale(c.ScaledHeight, float64(c.TargetWidth)/float64(c.TargetHeight))
			c.ResultCropHeight = c.ScaledHeight

		default:
			c.ResultCropWidth = c.TargetWidth
			c.ResultCropHeight = c.TargetHeight
		}
	} else {
		c.ResultCropWidth = c.TargetWidth
		c.ResultCropHeight = c.TargetHeight
	}

	if po.ExtendAspectRatioEnabled() && c.TargetWidth > 0 && c.TargetHeight > 0 {
		outWidth := imath.MinNonZero(c.ScaledWidth, c.ResultCropWidth)
		outHeight := imath.MinNonZero(c.ScaledHeight, c.ResultCropHeight)

		diffW := float64(c.TargetWidth) / float64(outWidth)
		diffH := float64(c.TargetHeight) / float64(outHeight)

		switch {
		case diffH > diffW:
			c.ExtendAspectRatioHeight = imath.Scale(outWidth, float64(c.TargetHeight)/float64(c.TargetWidth))
			c.ExtendAspectRatioWidth = outWidth

		case diffW > diffH:
			c.ExtendAspectRatioWidth = imath.Scale(outHeight, float64(c.TargetWidth)/float64(c.TargetHeight))
			c.ExtendAspectRatioHeight = outHeight
		}
	}
}

func (c *Context) limitScale(widthToScale, heightToScale int, po ProcessingOptions) {
	maxresultDim := c.PO.MaxResultDimension()

	if maxresultDim <= 0 {
		return
	}

	outWidth := imath.MinNonZero(c.ScaledWidth, c.ResultCropWidth)
	outHeight := imath.MinNonZero(c.ScaledHeight, c.ResultCropHeight)

	if po.ExtendEnabled() {
		outWidth = max(outWidth, c.TargetWidth)
		outHeight = max(outHeight, c.TargetHeight)
	} else if po.ExtendAspectRatioEnabled() {
		outWidth = max(outWidth, c.ExtendAspectRatioWidth)
		outHeight = max(outHeight, c.ExtendAspectRatioHeight)
	}

	outWidth += imath.ScaleToEven(po.PaddingLeft(), c.DprScale)
	outWidth += imath.ScaleToEven(po.PaddingRight(), c.DprScale)
	outHeight += imath.ScaleToEven(po.PaddingTop(), c.DprScale)
	outHeight += imath.ScaleToEven(po.PaddingBottom(), c.DprScale)

	if maxresultDim > 0 && (outWidth > maxresultDim || outHeight > maxresultDim) {
		downScale := float64(maxresultDim) / float64(max(outWidth, outHeight))

		c.WScale *= downScale
		c.HScale *= downScale

		// Prevent scaling below 1px
		if minWScale := 1.0 / float64(widthToScale); c.WScale < minWScale {
			c.WScale = minWScale
		}
		if minHScale := 1.0 / float64(heightToScale); c.HScale < minHScale {
			c.HScale = minHScale
		}

		c.DprScale *= downScale

		// Recalculate the sizes after changing the scales
		c.calcSizes(widthToScale, heightToScale, po)
	}
}

// CalcParams calculates context image parameters based on the current image size.
// Some steps (like trim) must call this function when finished.
func (c *Context) CalcParams() {
	c.SrcWidth, c.SrcHeight, c.Angle, c.Flip = ExtractGeometry(c.Img, c.PO.Rotate(), c.PO.AutoRotate())

	c.CropWidth = CalcCropSize(c.SrcWidth, c.PO.CropWidth())
	c.CropHeight = CalcCropSize(c.SrcHeight, c.PO.CropHeight())

	widthToScale := imath.MinNonZero(c.CropWidth, c.SrcWidth)
	heightToScale := imath.MinNonZero(c.CropHeight, c.SrcHeight)

	c.calcScale(widthToScale, heightToScale, c.PO)
	c.calcSizes(widthToScale, heightToScale, c.PO)
	c.limitScale(widthToScale, heightToScale, c.PO)
}
