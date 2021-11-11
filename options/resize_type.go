package options

import "fmt"

type ResizeType int

const (
	ResizeFit ResizeType = iota
	ResizeFill
	ResizeFillDown
	ResizeForce
	ResizeAuto
)

var resizeTypes = map[string]ResizeType{
	"fit":       ResizeFit,
	"fill":      ResizeFill,
	"fill-down": ResizeFillDown,
	"force":     ResizeForce,
	"auto":      ResizeAuto,
}

func (rt ResizeType) String() string {
	for k, v := range resizeTypes {
		if v == rt {
			return k
		}
	}
	return ""
}

func (rt ResizeType) MarshalJSON() ([]byte, error) {
	for k, v := range resizeTypes {
		if v == rt {
			return []byte(fmt.Sprintf("%q", k)), nil
		}
	}
	return []byte("null"), nil
}
