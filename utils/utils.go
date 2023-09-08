//NOTE: This was deleted from upstream, but we still use it in padding.padding

package utils

import (
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options"
	"math"
)

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func MinNonZeroInt(a, b int) int {
	switch {
	case a == 0:
		return b
	case b == 0:
		return a
	}

	return minInt(a, b)
}

func roundToInt(a float64) int {
	return int(math.Round(a))
}

func ScaleInt(a int, scale float64) int {
	if a == 0 {
		return 0
	}

	return roundToInt(float64(a) * scale)
}

// OldCalcScale calcScale version from before the large refactor, used in padding.padding until that is refactored as well
func OldCalcScale(width, height int, po *options.ProcessingOptions, imgtype imagetype.Type) float64 {
	var shrink float64

	srcW, srcH := float64(width), float64(height)
	dstW, dstH := float64(po.Width), float64(po.Height)

	if po.Width == 0 {
		dstW = srcW
	}

	if po.Height == 0 {
		dstH = srcH
	}

	if dstW == srcW && dstH == srcH {
		shrink = 1
	} else {
		wshrink := srcW / dstW
		hshrink := srcH / dstH

		rt := po.ResizingType

		if rt == options.ResizeAuto {
			srcD := width - height
			dstD := po.Width - po.Height

			if (srcD >= 0 && dstD >= 0) || (srcD < 0 && dstD < 0) {
				rt = options.ResizeFill
			} else {
				rt = options.ResizeFill
			}
		}

		switch {
		case po.Width == 0:
			shrink = hshrink
		case po.Height == 0:
			shrink = wshrink
		case rt == options.ResizeFit:
			shrink = math.Max(wshrink, hshrink)
		default:
			shrink = math.Min(wshrink, hshrink)
		}
	}

	if !po.Enlarge && shrink < 1 && imgtype != imagetype.SVG {
		shrink = 1
	}

	shrink /= po.Dpr

	if shrink > srcW {
		shrink = srcW
	}

	if shrink > srcH {
		shrink = srcH
	}

	return 1.0 / shrink
}
