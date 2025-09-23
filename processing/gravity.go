package processing

import (
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
)

var gravityTypesRotationMap = map[int]map[options.GravityType]options.GravityType{
	90: {
		options.GravityNorth:     options.GravityWest,
		options.GravityEast:      options.GravityNorth,
		options.GravitySouth:     options.GravityEast,
		options.GravityWest:      options.GravitySouth,
		options.GravityNorthWest: options.GravitySouthWest,
		options.GravityNorthEast: options.GravityNorthWest,
		options.GravitySouthWest: options.GravitySouthEast,
		options.GravitySouthEast: options.GravityNorthEast,
	},
	180: {
		options.GravityNorth:     options.GravitySouth,
		options.GravityEast:      options.GravityWest,
		options.GravitySouth:     options.GravityNorth,
		options.GravityWest:      options.GravityEast,
		options.GravityNorthWest: options.GravitySouthEast,
		options.GravityNorthEast: options.GravitySouthWest,
		options.GravitySouthWest: options.GravityNorthEast,
		options.GravitySouthEast: options.GravityNorthWest,
	},
	270: {
		options.GravityNorth:     options.GravityEast,
		options.GravityEast:      options.GravitySouth,
		options.GravitySouth:     options.GravityWest,
		options.GravityWest:      options.GravityNorth,
		options.GravityNorthWest: options.GravityNorthEast,
		options.GravityNorthEast: options.GravitySouthEast,
		options.GravitySouthWest: options.GravityNorthWest,
		options.GravitySouthEast: options.GravitySouthWest,
	},
}

var gravityTypesFlipMap = map[options.GravityType]options.GravityType{
	options.GravityEast:      options.GravityWest,
	options.GravityWest:      options.GravityEast,
	options.GravityNorthWest: options.GravityNorthEast,
	options.GravityNorthEast: options.GravityNorthWest,
	options.GravitySouthWest: options.GravitySouthEast,
	options.GravitySouthEast: options.GravitySouthWest,
}

type GravityOptions struct {
	Type options.GravityType
	X, Y float64
}

// NewGravityOptions builds a new [GravityOptions] instance.
// It fills the [GravityOptions] struct with the options values under the given prefix.
// If the gravity type is not set in the options,
// it returns a [GravityOptions] with the provided default type.
func NewGravityOptions(o ProcessingOptions, prefix string, defType options.GravityType) GravityOptions {
	gr := GravityOptions{
		Type: options.Get(o.Options, prefix+keys.SuffixType, defType),
		X:    o.GetFloat(prefix+keys.SuffixXOffset, 0.0),
		Y:    o.GetFloat(prefix+keys.SuffixYOffset, 0.0),
	}

	return gr
}

func (g *GravityOptions) RotateAndFlip(angle int, flip bool) {
	angle %= 360

	if flip {
		if gt, ok := gravityTypesFlipMap[g.Type]; ok {
			g.Type = gt
		}

		switch g.Type {
		case options.GravityCenter, options.GravityNorth, options.GravitySouth:
			g.X = -g.X
		case options.GravityFocusPoint:
			g.X = 1.0 - g.X
		}
	}

	if angle > 0 {
		if rotMap := gravityTypesRotationMap[angle]; rotMap != nil {
			if gt, ok := rotMap[g.Type]; ok {
				g.Type = gt
			}

			switch angle {
			case 90:
				switch g.Type {
				case options.GravityCenter, options.GravityEast, options.GravityWest:
					g.X, g.Y = g.Y, -g.X
				case options.GravityFocusPoint:
					g.X, g.Y = g.Y, 1.0-g.X
				default:
					g.X, g.Y = g.Y, g.X
				}
			case 180:
				switch g.Type {
				case options.GravityCenter:
					g.X, g.Y = -g.X, -g.Y
				case options.GravityNorth, options.GravitySouth:
					g.X = -g.X
				case options.GravityEast, options.GravityWest:
					g.Y = -g.Y
				case options.GravityFocusPoint:
					g.X, g.Y = 1.0-g.X, 1.0-g.Y
				}
			case 270:
				switch g.Type {
				case options.GravityCenter, options.GravityNorth, options.GravitySouth:
					g.X, g.Y = -g.Y, g.X
				case options.GravityFocusPoint:
					g.X, g.Y = 1.0-g.Y, g.X
				default:
					g.X, g.Y = g.Y, g.X
				}
			}
		}
	}
}
