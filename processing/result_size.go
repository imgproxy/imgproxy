package processing

import (
	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
)

func resultSize(po *options.ProcessingOptions) (int, int) {
	resultWidth := imath.Scale(po.Width, po.Dpr*po.ZoomWidth)
	resultHeight := imath.Scale(po.Height, po.Dpr*po.ZoomHeight)

	return resultWidth, resultHeight
}
