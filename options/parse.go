package options

import (
	"fmt"
	"slices"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type number = interface {
	~int | ~int64 | ~float32 | ~float64
}

func parseNumber[T number](arg string) (T, error) {
	var zero T
	var val T
	var err error

	switch any(zero).(type) {
	case int:
		v, parseErr := strconv.Atoi(arg)
		val, err = T(v), parseErr
	case int64:
		v, parseErr := strconv.ParseInt(arg, 10, 64)
		val, err = T(v), parseErr
	case float32:
		v, parseErr := strconv.ParseFloat(arg, 32)
		val, err = T(v), parseErr
	case float64:
		v, parseErr := strconv.ParseFloat(arg, 64)
		val, err = T(v), parseErr
	}

	if err != nil {
		return zero, fmt.Errorf("invalid value: %s", arg)
	}

	return val, nil
}

// parseBool parses a string into a boolean value, logging a warning if the value is invalid
func parseBool(str string) bool {
	b, err := strconv.ParseBool(str)

	if err != nil {
		log.Warningf("`%s` is not a valid boolean value. Treated as false", str)
	}

	return b
}

func isGravityOffsetValid(gravity GravityType, offset float64) bool {
	return gravity != GravityFocusPoint || (offset >= 0 && offset <= 1)
}

// parseGravity parses gravity options from the provided arguments
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
			return newOptionArgumentError("Invalid %s arguments: %v", name, args)
		}
		g.X, g.Y = 0.0, 0.0
		return nil

	case GravityFocusPoint:
		if nArgs != 3 {
			return newOptionArgumentError("Invalid %s arguments: %v", name, args)
		}

	default:
		if nArgs > 3 {
			return newOptionArgumentError("Invalid %s arguments: %v", name, args)
		}
	}

	if nArgs > 1 {
		if x, err := strconv.ParseFloat(args[1], 64); err == nil && isGravityOffsetValid(g.Type, x) {
			g.X = x
		} else {
			return newOptionArgumentError("Invalid %s X: %s", name, args[1])
		}
	}

	if nArgs > 2 {
		if y, err := strconv.ParseFloat(args[2], 64); err == nil && isGravityOffsetValid(g.Type, y) {
			g.Y = y
		} else {
			return newOptionArgumentError("Invalid %s Y: %s", name, args[2])
		}
	}

	return nil
}

// parseExtend parses extend options from the provided arguments
func parseExtend(opts *ExtendOptions, name string, args []string) error {
	if len(args) > 4 {
		return newOptionArgumentError("Invalid %s arguments: %v", name, args)
	}

	opts.Enabled = parseBool(args[0])

	if len(args) > 1 {
		return parseGravity(&opts.Gravity, name+" gravity", args[1:], extendGravityTypes)
	}

	return nil
}

func parseDimension(d *int, name, arg string) error {
	if v, err := strconv.Atoi(arg); err == nil && v >= 0 {
		*d = v
	} else {
		return newOptionArgumentError("Invalid %s: %s", name, arg)
	}

	return nil
}
