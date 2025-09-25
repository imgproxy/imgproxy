package processing

import (
	"math"

	"github.com/imgproxy/imgproxy/v3/imath"
)

func calcPosition(width, height, innerWidth, innerHeight int, gravity *GravityOptions, dpr float64, allowOverflow bool) (left, top int) {
	if gravity.Type == GravityFocusPoint {
		pointX := imath.ScaleToEven(width, gravity.X)
		pointY := imath.ScaleToEven(height, gravity.Y)

		left = pointX - innerWidth/2
		top = pointY - innerHeight/2
	} else {
		var offX, offY int

		if math.Abs(gravity.X) >= 1.0 {
			offX = imath.RoundToEven(gravity.X * dpr)
		} else {
			offX = imath.ScaleToEven(width, gravity.X)
		}

		if math.Abs(gravity.Y) >= 1.0 {
			offY = imath.RoundToEven(gravity.Y * dpr)
		} else {
			offY = imath.ScaleToEven(height, gravity.Y)
		}

		left = imath.ShrinkToEven(width-innerWidth+1, 2) + offX
		top = imath.ShrinkToEven(height-innerHeight+1, 2) + offY

		if gravity.Type == GravityNorth || gravity.Type == GravityNorthEast || gravity.Type == GravityNorthWest {
			top = 0 + offY
		}

		if gravity.Type == GravityEast || gravity.Type == GravityNorthEast || gravity.Type == GravitySouthEast {
			left = width - innerWidth - offX
		}

		if gravity.Type == GravitySouth || gravity.Type == GravitySouthEast || gravity.Type == GravitySouthWest {
			top = height - innerHeight - offY
		}

		if gravity.Type == GravityWest || gravity.Type == GravityNorthWest || gravity.Type == GravitySouthWest {
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

	left = max(minX, min(left, maxX))
	top = max(minY, min(top, maxY))

	return
}
