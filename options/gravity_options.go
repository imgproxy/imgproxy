package options

import (
	"fmt"
)

type GravityType int

const (
	GravityUnknown GravityType = iota
	GravityCenter
	GravityNorth
	GravityEast
	GravitySouth
	GravityWest
	GravityNorthWest
	GravityNorthEast
	GravitySouthWest
	GravitySouthEast
	GravitySmart
	GravityFocusPoint
)

var gravityTypes = map[string]GravityType{
	"ce":   GravityCenter,
	"no":   GravityNorth,
	"ea":   GravityEast,
	"so":   GravitySouth,
	"we":   GravityWest,
	"nowe": GravityNorthWest,
	"noea": GravityNorthEast,
	"sowe": GravitySouthWest,
	"soea": GravitySouthEast,
	"sm":   GravitySmart,
	"fp":   GravityFocusPoint,
}

var gravityTypesRotationMap = map[int]map[GravityType]GravityType{
	90: {
		GravityNorth:     GravityWest,
		GravityEast:      GravityNorth,
		GravitySouth:     GravityEast,
		GravityWest:      GravitySouth,
		GravityNorthWest: GravitySouthWest,
		GravityNorthEast: GravityNorthWest,
		GravitySouthWest: GravitySouthEast,
		GravitySouthEast: GravityNorthEast,
	},
	180: {
		GravityNorth:     GravitySouth,
		GravityEast:      GravityWest,
		GravitySouth:     GravityNorth,
		GravityWest:      GravityEast,
		GravityNorthWest: GravitySouthEast,
		GravityNorthEast: GravitySouthWest,
		GravitySouthWest: GravityNorthEast,
		GravitySouthEast: GravityNorthWest,
	},
	270: {
		GravityNorth:     GravityEast,
		GravityEast:      GravitySouth,
		GravitySouth:     GravityWest,
		GravityWest:      GravityNorth,
		GravityNorthWest: GravityNorthEast,
		GravityNorthEast: GravitySouthEast,
		GravitySouthWest: GravityNorthWest,
		GravitySouthEast: GravitySouthWest,
	},
}

var gravityTypesFlipMap = map[GravityType]GravityType{
	GravityEast:      GravityWest,
	GravityWest:      GravityEast,
	GravityNorthWest: GravityNorthEast,
	GravityNorthEast: GravityNorthWest,
	GravitySouthWest: GravitySouthEast,
	GravitySouthEast: GravitySouthWest,
}

func (gt GravityType) String() string {
	for k, v := range gravityTypes {
		if v == gt {
			return k
		}
	}
	return ""
}

func (gt GravityType) MarshalJSON() ([]byte, error) {
	for k, v := range gravityTypes {
		if v == gt {
			return []byte(fmt.Sprintf("%q", k)), nil
		}
	}
	return []byte("null"), nil
}

type GravityOptions struct {
	Type GravityType
	X, Y float64
}

func (g *GravityOptions) RotateAndFlip(angle int, flip bool) {
	angle %= 360

	if flip {
		if gt, ok := gravityTypesFlipMap[g.Type]; ok {
			g.Type = gt
		}

		switch g.Type {
		case GravityCenter, GravityNorth, GravitySouth:
			g.X = -g.X
		case GravityFocusPoint:
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
				case GravityCenter, GravityEast, GravityWest:
					g.X, g.Y = g.Y, -g.X
				case GravityFocusPoint:
					g.X, g.Y = g.Y, 1.0-g.X
				default:
					g.X, g.Y = g.Y, g.X
				}
			case 180:
				switch g.Type {
				case GravityCenter:
					g.X, g.Y = -g.X, -g.Y
				case GravityNorth, GravitySouth:
					g.X = -g.X
				case GravityEast, GravityWest:
					g.Y = -g.Y
				case GravityFocusPoint:
					g.X, g.Y = 1.0-g.X, 1.0-g.Y
				}
			case 270:
				switch g.Type {
				case GravityCenter, GravityNorth, GravitySouth:
					g.X, g.Y = -g.Y, g.X
				case GravityFocusPoint:
					g.X, g.Y = 1.0-g.Y, g.X
				default:
					g.X, g.Y = g.Y, g.X
				}
			}
		}
	}
}
