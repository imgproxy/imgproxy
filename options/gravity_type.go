package options

import "fmt"

type GravityType int

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

	// Watermark gravity types
	GravityReplicate
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
	"re":   GravityReplicate,
}

var commonGravityTypes = []GravityType{
	GravityCenter,
	GravityNorth,
	GravityEast,
	GravitySouth,
	GravityWest,
	GravityNorthWest,
	GravityNorthEast,
	GravitySouthWest,
	GravitySouthEast,
}

var cropGravityTypes = append(
	[]GravityType{
		GravitySmart,
		GravityFocusPoint,
	},
	commonGravityTypes...,
)

var extendGravityTypes = append(
	[]GravityType{
		GravityFocusPoint,
	},
	commonGravityTypes...,
)

var watermarkGravityTypes = append(
	[]GravityType{
		GravityReplicate,
	},
	commonGravityTypes...,
)
