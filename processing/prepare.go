package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/vips"
)

// ExtractGeometry extracts image width, height, orientation angle and flip flag from the image metadata.
func (c *Context) ExtractGeometry(img *vips.Image, baseAngle int, autoRotate bool) (int, int, int, bool) {
	width := img.Width()
	height := img.Height()

	angle, flip := c.angleFlip(img, autoRotate)

	if (angle+baseAngle)%180 != 0 {
		width, height = height, width
	}

	return width, height, angle, flip
}

// angleFlip returns the orientation angle and flip flag based on the image metadata
// and po.AutoRotate flag.
func (c *Context) angleFlip(img *vips.Image, autoRotate bool) (int, bool) {
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
func (c *Context) CalcCropSize(orig int, crop float64) int {
	switch {
	case crop == 0.0:
		return 0
	case crop >= 1.0:
		return int(crop)
	default:
		return max(1, imath.Scale(orig, crop))
	}
}

// calcShrink calculates the destination size and shrink factor
func calcShrink(value int, src, dst float64) (float64, float64) {
	if value == 0 {
		dst = src
	}

	shrink := 1.0
	if dst != src {
		shrink = src / dst
	}

	return dst, shrink
}

func (c *Context) calcScale(width, height int, po *options.ProcessingOptions) {
	var wshrink, hshrink float64

	srcW, srcH := float64(width), float64(height)
	dstW, dstH := float64(po.Width), float64(po.Height)

	dstW, wshrink = calcShrink(po.Width, srcW, dstW)
	dstH, hshrink = calcShrink(po.Height, srcH, dstH)

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

	c.DprScale = po.Dpr

	if !po.Enlarge && c.ImgData != nil && !c.ImgData.Format().IsVector() {
		minShrink := math.Min(wshrink, hshrink)
		if minShrink < 1 {
			wshrink /= minShrink
			hshrink /= minShrink

			if !po.Extend.Enabled {
				c.DprScale /= minShrink
			}
		}

		// The minimum of wshrink and hshrink is the maximum dprScale value
		// that can be used without enlarging the image.
		c.DprScale = math.Min(c.DprScale, math.Min(wshrink, hshrink))
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

func (c *Context) calcSizes(widthToScale, heightToScale int, po *options.ProcessingOptions) {
	c.TargetWidth = imath.Scale(po.Width, c.DprScale*po.ZoomWidth)
	c.TargetHeight = imath.Scale(po.Height, c.DprScale*po.ZoomHeight)

	c.ScaledWidth = imath.Scale(widthToScale, c.WScale)
	c.ScaledHeight = imath.Scale(heightToScale, c.HScale)

	if po.ResizingType == options.ResizeFillDown && !po.Enlarge {
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

	if po.ExtendAspectRatio.Enabled && c.TargetWidth > 0 && c.TargetHeight > 0 {
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

func (c *Context) limitScale(widthToScale, heightToScale int, po *options.ProcessingOptions) {
	maxresultDim := po.SecurityOptions.MaxResultDimension

	if maxresultDim <= 0 {
		return
	}

	outWidth := imath.MinNonZero(c.ScaledWidth, c.ResultCropWidth)
	outHeight := imath.MinNonZero(c.ScaledHeight, c.ResultCropHeight)

	if po.Extend.Enabled {
		outWidth = max(outWidth, c.TargetWidth)
		outHeight = max(outHeight, c.TargetHeight)
	} else if po.ExtendAspectRatio.Enabled {
		outWidth = max(outWidth, c.ExtendAspectRatioWidth)
		outHeight = max(outHeight, c.ExtendAspectRatioHeight)
	}

	if po.Padding.Enabled {
		outWidth += imath.ScaleToEven(po.Padding.Left, c.DprScale) + imath.ScaleToEven(po.Padding.Right, c.DprScale)
		outHeight += imath.ScaleToEven(po.Padding.Top, c.DprScale) + imath.ScaleToEven(po.Padding.Bottom, c.DprScale)
	}

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

// Prepare calculates context image parameters based on the current image size.
// Some steps (like trim) must call this function when finished.
func (c *Context) CalcParams() {
	if c.ImgData == nil {
		return
	}

	c.SrcWidth, c.SrcHeight, c.Angle, c.Flip = c.ExtractGeometry(c.Img, c.PO.Rotate, c.PO.AutoRotate)

	c.CropWidth = c.CalcCropSize(c.SrcWidth, c.PO.Crop.Width)
	c.CropHeight = c.CalcCropSize(c.SrcHeight, c.PO.Crop.Height)

	widthToScale := imath.MinNonZero(c.CropWidth, c.SrcWidth)
	heightToScale := imath.MinNonZero(c.CropHeight, c.SrcHeight)

	c.calcScale(widthToScale, heightToScale, c.PO)
	c.calcSizes(widthToScale, heightToScale, c.PO)
	c.limitScale(widthToScale, heightToScale, c.PO)
}
