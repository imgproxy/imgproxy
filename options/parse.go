package options

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"
)

// ensureMaxArgs checks if the number of arguments is as expected
func ensureMaxArgs(name string, args []string, max int) error {
	if len(args) > max {
		return newInvalidArgsError(name, args)
	}
	return nil
}

// parseBool parses a boolean option value and warns if the value is invalid
func parseBool(value *bool, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	b, err := strconv.ParseBool(args[0])

	if err != nil {
		slog.Warn(fmt.Sprintf("%s `%s` is not a valid boolean value. Treated as false", name, args[0]))
	}

	*value = b
	return nil
}

// parseFloat64 parses a float64 option value
func parseFloat64(value *float64, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil {
		return newInvalidArgsError(name, args)
	}

	*value = f
	return nil
}

// parsePositiveFloat64 parses a positive float64 option value
func parsePositiveFloat64(value *float64, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f < 0 {
		return newInvalidArgsError(name, args, "positive number or 0")
	}
	*value = f
	return nil
}

// parsePositiveFloat64 parses a positive float64 option value
func parsePositiveNonZeroFloat64(value *float64, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 64)
	if err != nil || f <= 0 {
		return newInvalidArgsError(name, args, "positive number")
	}
	*value = f
	return nil
}

// parsePositiveFloat32 parses a positive float32 option value
func parsePositiveNonZeroFloat32(value *float32, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	f, err := strconv.ParseFloat(args[0], 32)
	if err != nil || f <= 0 {
		return newInvalidArgsError(name, args, "positive number")
	}
	*value = float32(f)
	return nil
}

// parseInt parses a positive integer option value
func parseInt(value *int, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil {
		return newOptionArgumentError(name, args)
	}
	*value = i
	return nil
}

// parsePositiveNonZeroInt parses a positive non-zero integer option value
func parsePositiveNonZeroInt(value *int, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i <= 0 {
		return newInvalidArgsError(name, args, "positive number")
	}
	*value = i
	return nil
}

// parsePositiveInt parses a positive integer option value
func parsePositiveInt(value *int, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i < 0 {
		return newOptionArgumentError("Invalid %s arguments: %s (expected positive number)", name, args)
	}
	*value = i
	return nil
}

// parseQualityInt parses a quality integer option value (1-100)
func parseQualityInt(value *int, name string, args ...string) error {
	if err := ensureMaxArgs(name, args, 1); err != nil {
		return err
	}

	i, err := strconv.Atoi(args[0])
	if err != nil || i < 1 || i > 100 {
		return newInvalidArgsError(name, args, "number in range 1-100")
	}
	*value = i
	return nil
}

func isGravityOffcetValid(gravity GravityType, offset float64) bool {
	return gravity != GravityFocusPoint || (offset >= 0 && offset <= 1)
}

func parseGravity(g *GravityOptions, name string, args []string, allowedTypes []GravityType) error {
	nArgs := len(args)

	if t, ok := gravityTypes[args[0]]; ok && slices.Contains(allowedTypes, t) {
		g.Type = t
	} else {
		return newOptionArgumentError("Invalid %s: %s", name, args[0])
	}

	switch g.Type {
	case GravitySmart:
		if nArgs > 1 {
			return newInvalidArgsError(name, args)
		}
		g.X, g.Y = 0.0, 0.0

	case GravityFocusPoint:
		if nArgs != 3 {
			return newInvalidArgsError(name, args)
		}
		fallthrough

	default:
		if nArgs > 3 {
			return newInvalidArgsError(name, args)
		}

		if nArgs > 1 {
			if x, err := strconv.ParseFloat(args[1], 64); err == nil && isGravityOffcetValid(g.Type, x) {
				g.X = x
			} else {
				return newOptionArgumentError("Invalid %s X: %s", name, args[1])
			}
		}

		if nArgs > 2 {
			if y, err := strconv.ParseFloat(args[2], 64); err == nil && isGravityOffcetValid(g.Type, y) {
				g.Y = y
			} else {
				return newOptionArgumentError("Invalid %s Y: %s", name, args[2])
			}
		}
	}

	return nil
}

func parseExtend(opts *ExtendOptions, name string, args []string) error {
	if err := ensureMaxArgs(name, args, 4); err != nil {
		return err
	}

	if err := parseBool(&opts.Enabled, name+" enabled", args[0]); err != nil {
		return err
	}

	if len(args) > 1 {
		return parseGravity(&opts.Gravity, name+" gravity", args[1:], extendGravityTypes)
	}

	return nil
}
