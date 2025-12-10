package processing

import "fmt"

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
	for k, v := range ResizeTypes {
		if v == rt {
			return k
		}
	}
	return ""
}

func (rt ResizeType) MarshalJSON() ([]byte, error) {
	for k, v := range ResizeTypes {
		if v == rt {
			return fmt.Appendf([]byte{}, "%q", k), nil
		}
	}
	return []byte("null"), nil
}
