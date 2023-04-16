package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v3/imath"
	"github.com/imgproxy/imgproxy/v3/options"
)

func calcPosition(width, height, innerWidth, innerHeight int, gravity *options.GravityOptions, dpr float64, allowOverflow bool) (left, top int) {
	if gravity.Type == options.GravityFocusPoint {
		pointX := imath.ScaleToEven(width, gravity.X)
		pointY := imath.ScaleToEven(height, gravity.Y)

		left = pointX - innerWidth/2
		top = pointY - innerHeight/2
	} else {
		offX, offY := int(math.RoundToEven(gravity.X*dpr)), int(math.RoundToEven(gravity.Y*dpr))

		left = imath.ShrinkToEven(width-innerWidth+1, 2) + offX
		top = imath.ShrinkToEven(height-innerHeight+1, 2) + offY

		if gravity.Type == options.GravityNorth || gravity.Type == options.GravityNorthEast || gravity.Type == options.GravityNorthWest {
			top = 0 + offY
		}

		if gravity.Type == options.GravityEast || gravity.Type == options.GravityNorthEast || gravity.Type == options.GravitySouthEast {
			left = width - innerWidth - offX
		}

		if gravity.Type == options.GravitySouth || gravity.Type == options.GravitySouthEast || gravity.Type == options.GravitySouthWest {
			top = height - innerHeight - offY
		}

		if gravity.Type == options.GravityWest || gravity.Type == options.GravityNorthWest || gravity.Type == options.GravitySouthWest {
			left = 0 + offX
		}
	}

	var minX, maxX, minY, maxY int

	if allowOverflow {
		minX, maxX = -innerWidth+1, width-1
		minY, maxY = -innerHeight+1, height-1
	} else {
		minX, maxX = 0, width-innerWidth
		minY, maxY = 0, height-innerHeight
	}

	left = imath.Max(minX, imath.Min(left, maxX))
	top = imath.Max(minY, imath.Min(top, maxY))

	return
}
