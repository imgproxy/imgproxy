package optionsparser

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/processing"
	"github.com/imgproxy/imgproxy/v3/vips/color"
)

// ensureMaxArgs checks if the number of arguments is as expected
func ensureMaxArgs(name string, args []string, max int) error {
	if len(args) > max {
		return newInvalidArgsError(name, args)
	}
	return nil
}

// parseBool parses a boolean option value and warns if the value is invalid
func parseBool(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	b, err := strconv.ParseBool(args[0])

	if err != nil {
		slog.Warn(fmt.Sprintf("%s `%s` is not a valid boolean value. Treated as false", key, args[0]))
	}

	o.Set(key, b)

	return nil
}

// parseFloat parses a float64 option value
func parseFloat(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return newInvalidArgsError(key, args)
	}

	o.Set(key, f)

	return nil
}

// parsePositiveFloat parses a positive float64 option value
func parsePositiveFloat(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f < 0 {
		return newInvalidArgumentError(key, args[0], "positive number or 0")
	}

	o.Set(key, f)

	return nil
}

// parsePositiveNonZeroFloat parses a positive non-zero float64 option value
func parsePositiveNonZeroFloat(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f <= 0 {
		return newInvalidArgumentError(key, args[0], "positive number")
	}

	o.Set(key, f)

	return nil
}

// parseInt parses a positive integer option value
func parseInt(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil {
		return newInvalidArgumentError(key, args[0], "integer number")
	}

	o.Set(key, i)

	return nil
}

// parsePositiveNonZeroInt parses a positive non-zero integer option value
func parsePositiveNonZeroInt(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i <= 0 {
		return newInvalidArgumentError(key, args[0], "positive number")
	}

	o.Set(key, i)

	return nil
}

// parsePositiveInt parses a positive integer option value
func parsePositiveInt(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i < 0 {
		return newInvalidArgumentError(key, args[0], "positive number or 0")
	}

	o.Set(key, i)

	return nil
}

// parseQualityInt parses a quality integer option value (1-100)
func parseQualityInt(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i < 1 || i > 100 {
		return newInvalidArgumentError(key, args[0], "number in range 1-100")
	}

	o.Set(key, i)

	return nil
}

// parseOpacityFloat parses an opacity float option value (0-1)
func parseOpacityFloat(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f < 0 || f > 1 {
		return newInvalidArgumentError(key, args[0], "number in range 0-1")
	}

	o.Set(key, f)

	return nil
}

// parseResolution parses a resolution option value in megapixels and stores it as pixels
func parseResolution(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f < 0 {
		return newInvalidArgumentError(key, args[0], "positive number or 0")
	}

	// Resolution is defined as megapixels but stored as pixels
	o.Set(key, int(f*1000000))

	return nil
}

// parseBase64String parses a base64-encoded string option value
func parseBase64String(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	b, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(args[0], "="))
	if err != nil {
		return newInvalidArgumentError(key, args[0], "URL-safe base64-encoded string")
	}

	o.Set(key, string(b))

	return nil
}

// parseHexRGBColor parses a hex-encoded RGB color option value
func parseHexRGBColor(o *options.Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	c, err := color.RGBFromHex(args[0])
	if err != nil {
		return newInvalidArgumentError(key, args[0], "hex-encoded color")
	}

	o.Set(key, c)

	return nil
}

// parseFromMap parses an option value from a map of allowed values
func parseFromMap[T comparable](o *options.Options, key string, m map[string]T, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	v, ok := m[args[0]]
	if !ok {
		return newInvalidArgumentError(key, args[0], slices.Collect(maps.Keys(m))...)
	}

	o.Set(key, v)

	return nil
}

func parseGravityType(
	o *options.Options,
	key string,
	allowedTypes []processing.GravityType,
	args ...string,
) (processing.GravityType, error) {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return processing.GravityUnknown, err
	}

	gType, ok := processing.GravityTypes[args[0]]
	if !ok || !slices.Contains(allowedTypes, gType) {
		types := make([]string, len(allowedTypes))
		for i, at := range allowedTypes {
			types[i] = at.String()
		}
		return processing.GravityUnknown, newInvalidArgumentError(key, args[0], types...)
	}

	o.Set(key, gType)

	return gType, nil
}

func isGravityOffsetValid(gravity processing.GravityType, offset float64) bool {
	return gravity != processing.GravityFocusPoint || (offset >= 0 && offset <= 1)
}

func parseGravity(
	o *options.Options,
	key string,
	allowedTypes []processing.GravityType,
	args ...string,
) error {
	nArgs := len(args)

	keyType := key + keys.SuffixType
	keyXOffset := key + keys.SuffixXOffset
	keyYOffset := key + keys.SuffixYOffset

	gType, err := parseGravityType(o, keyType, allowedTypes, args[0])
	if err != nil {
		return err
	}

	switch gType {
	case processing.GravitySmart:
		if nArgs > 1 {
			return newInvalidArgsError(key, args)
		}
		o.Delete(keyXOffset)
		o.Delete(keyYOffset)

	case processing.GravityFocusPoint:
		if nArgs != 3 {
			return newInvalidArgsError(key, args)
		}
		fallthrough

	default:
		if nArgs > 3 {
			return newInvalidArgsError(key, args)
		}

		if nArgs > 1 {
			if x, err := strconv.ParseFloat(args[1], 64); err == nil && isGravityOffsetValid(gType, x) {
				o.Set(keyXOffset, x)
			} else {
				return newInvalidArgumentError(keyXOffset, args[1])
			}
		}

		if nArgs > 2 {
			if y, err := strconv.ParseFloat(args[2], 64); err == nil && isGravityOffsetValid(gType, y) {
				o.Set(keyYOffset, y)
			} else {
				return newInvalidArgumentError(keyYOffset, args[2])
			}
		}
	}

	return nil
}

func parseExtend(o *options.Options, key string, args []string) error {
	if err := ensureMaxArgs(key, args, 4); err != nil {
		return err
	}

	if err := parseBool(o, key+keys.SuffixEnabled, args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		return parseGravity(o, key+keys.SuffixGravity, processing.ExtendGravityTypes, args[1:]...)
	}

	return nil
}
