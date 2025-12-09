package processing

import "fmt"

type GravityType int

func (gt GravityType) String() string {
	for k, v := range GravityTypes {
		if v == gt {
			return k
		}
	}
	return ""
}

func (gt GravityType) MarshalJSON() ([]byte, error) {
	for k, v := range GravityTypes {
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

	// GravityReplicate and below: watermark gravity types
	GravityReplicate
)

var GravityTypes = map[string]GravityType{
	"ce":   GravityCenter,
	"no":   GravityNorth,
	"ea":   GravityEast,
	"so":   GravitySouth,
	"we":   GravityWest,
	"nowe": GravityNorthWest, //nolint:misspell
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

var CropGravityTypes = append(
	[]GravityType{
		GravitySmart,
		GravityFocusPoint,
	},
	commonGravityTypes...,
)

var ExtendGravityTypes = append(
	[]GravityType{
		GravityFocusPoint,
	},
	commonGravityTypes...,
)

var WatermarkGravityTypes = append(
	[]GravityType{
		GravityReplicate,
	},
	commonGravityTypes...,
)
