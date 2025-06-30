package config

import "fmt"

type WebpPresetKind int

const (
	WebpPresetDefault WebpPresetKind = iota
	WebpPresetPhoto
	WebpPresetPicture
	WebpPresetDrawing
	WebpPresetIcon
	WebpPresetText
)

var WebpPresets = map[string]WebpPresetKind{
	"default": WebpPresetDefault,
	"photo":   WebpPresetPhoto,
	"picture": WebpPresetPicture,
	"drawing": WebpPresetDrawing,
	"icon":    WebpPresetIcon,
	"text":    WebpPresetText,
}

func (wp WebpPresetKind) String() string {
	for k, v := range WebpPresets {
		if v == wp {
			return k
		}
	}
	return ""
}

func (wp WebpPresetKind) MarshalJSON() ([]byte, error) {
	for k, v := range WebpPresets {
		if v == wp {
			return []byte(fmt.Sprintf("%q", k)), nil
		}
	}
	return []byte("null"), nil
}
