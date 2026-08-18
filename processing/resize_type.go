package processing

import (
	"fmt"

	"github.com/imgproxy/imgproxy/v4/env"
)

type ResizeType int

const (
	ResizeFit ResizeType = iota
	ResizeFill
	ResizeFillDown
	ResizeForce
	ResizeAuto
)

var ResizeTypes = map[string]ResizeType{
	"fit":       ResizeFit,
	"fill":      ResizeFill,
	"fill-down": ResizeFillDown,
	"force":     ResizeForce,
	"auto":      ResizeAuto,
}

func (rt ResizeType) String() string {
	return env.EnumName(ResizeTypes, rt, "")
}

func (rt ResizeType) MarshalJSON() ([]byte, error) {
	if s := rt.String(); s != "" {
		return fmt.Appendf([]byte{}, "%q", s), nil
	}
	return []byte("null"), nil
}
