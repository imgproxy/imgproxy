package processing

import (
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
)

func resultSize(po *options.ProcessingOptions, dprScale float64) (int, int) {
	resultWidth := imath.Scale(po.Width, dprScale*po.ZoomWidth)
	resultHeight := imath.Scale(po.Height, dprScale*po.ZoomHeight)

	return resultWidth, resultHeight
}
