package options

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"

	"github.com/imgproxy/imgproxy/v3/options/keys"
)

// ensureMaxArgs checks if the number of arguments is as expected
func ensureMaxArgs(name string, args []string, max int) error {
	if len(args) > max {
		return newInvalidArgsError(name, args)
	}
	return nil
}

// parseBool parses a boolean option value and warns if the value is invalid
func parseBool(o *Options, key string, args ...string) error {
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
func parseFloat(o *Options, key string, args ...string) error {
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
func parsePositiveFloat(o *Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f < 0 {
		return newInvalidArgsError(key, args, "positive number or 0")
	}

	o.Set(key, f)

	return nil
}

// parsePositiveNonZeroFloat parses a positive non-zero float64 option value
func parsePositiveNonZeroFloat(o *Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f <= 0 {
		return newInvalidArgsError(key, args, "positive number")
	}

	o.Set(key, f)

	return nil
}

// parseInt parses a positive integer option value
func parseInt(o *Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil {
		return newOptionArgumentError(key, args)
	}

	o.Set(key, i)

	return nil
}

// parsePositiveNonZeroInt parses a positive non-zero integer option value
func parsePositiveNonZeroInt(o *Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i <= 0 {
		return newInvalidArgsError(key, args, "positive number")
	}

	o.Set(key, i)

	return nil
}

// parsePositiveInt parses a positive integer option value
func parsePositiveInt(o *Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i < 0 {
		return newOptionArgumentError("Invalid %s arguments: %s (expected positive number)", key, args)
	}

	o.Set(key, i)

	return nil
}

// parseQualityInt parses a quality integer option value (1-100)
func parseQualityInt(o *Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i < 1 || i > 100 {
		return newInvalidArgsError(key, args, "number in range 1-100")
	}

	o.Set(key, i)

	return nil
}

// parseResolution parses a resolution option value in megapixels and stores it as pixels
func parseResolution(o *Options, key string, args ...string) error {
	if err := ensureMaxArgs(key, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f < 0 {
		return newInvalidArgsError(key, args, "positive number or 0")
	}

	// Resolution is defined as megapixels but stored as pixels
	o.Set(key, int(f*1000000))

	return nil
}

func isGravityOffcetValid(gravity GravityType, offset float64) bool {
	return gravity != GravityFocusPoint || (offset >= 0 && offset <= 1)
}

func parseGravity(
	o *Options,
	key string,
	args []string,
	allowedTypes []GravityType,
) error {
	nArgs := len(args)

	keyType := key + keys.SuffixType
	keyXOffset := key + keys.SuffixXOffset
	keyYOffset := key + keys.SuffixYOffset

	gType, ok := gravityTypes[args[0]]
	if ok && slices.Contains(allowedTypes, gType) {
		o.Set(keyType, gType)
	} else {
		return newOptionArgumentError("Invalid %s: %s", keyType, args[0])
	}

	switch gType {
	case GravitySmart:
		if nArgs > 1 {
			return newInvalidArgsError(key, args)
		}
		o.Delete(keyXOffset)
		o.Delete(keyYOffset)

	case GravityFocusPoint:
		if nArgs != 3 {
			return newInvalidArgsError(key, args)
		}
		fallthrough

	default:
		if nArgs > 3 {
			return newInvalidArgsError(key, args)
		}

		if nArgs > 1 {
			if x, err := strconv.ParseFloat(args[1], 64); err == nil && isGravityOffcetValid(gType, x) {
				o.Set(keyXOffset, x)
			} else {
				return newOptionArgumentError("Invalid %s: %s", keyXOffset, args[1])
			}
		}

		if nArgs > 2 {
			if y, err := strconv.ParseFloat(args[2], 64); err == nil && isGravityOffcetValid(gType, y) {
				o.Set(keyYOffset, y)
			} else {
				return newOptionArgumentError("Invalid %s: %s", keyYOffset, args[2])
			}
		}
	}

	return nil
}

func parseExtend(o *Options, key string, args []string) error {
	if err := ensureMaxArgs(key, args, 4); err != nil {
		return err
	}

	if err := parseBool(o, key+keys.SuffixEnabled, args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		return parseGravity(o, key+keys.SuffixGravity, args[1:], extendGravityTypes)
	}

	return nil
}
