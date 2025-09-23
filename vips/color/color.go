package color

import (
	"fmt"
	"log/slog"
	"regexp"
)

var hexColorRegex = regexp.MustCompile("^([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$")

const (
	hexColorLongFormat  = "%02x%02x%02x"
	hexColorShortFormat = "%1x%1x%1x"
)

type RGB struct{ R, G, B uint8 }

func RGBFromHex(hexcolor string) (RGB, error) {
	c := RGB{}

	if !hexColorRegex.MatchString(hexcolor) {
		return c, newColorError("Invalid hex color: %s", hexcolor)
	}

	if len(hexcolor) == 3 {
		fmt.Sscanf(hexcolor, hexColorShortFormat, &c.R, &c.G, &c.B)
		c.R *= 17
		c.G *= 17
		c.B *= 17
	} else {
		fmt.Sscanf(hexcolor, hexColorLongFormat, &c.R, &c.G, &c.B)
	}

	return c, nil
}

func (c RGB) LogValue() slog.Value {
	return slog.StringValue(fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B))
}
