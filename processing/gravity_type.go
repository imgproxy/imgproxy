package processing

import (
	"fmt"
	"log/slog"
)

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
			return fmt.Appendf([]byte{}, "%q", k), nil
		}
	}
	return fmt.Appendf([]byte{}, "%s", "null"), nil
}

func (gt GravityType) LogValue() slog.Value {
	return slog.StringValue(gt.String())
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
