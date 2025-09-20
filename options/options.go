package options

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/options/keys"
)

type Options map[string]any

// New creates a new Options map
func New() Options {
	return make(Options)
}

// Append appends a value to a slice option.
// If the option does not exist, it creates a new slice with the value.
func Append[T any](o Options, key string, value T) {
	if v, ok := o[key]; ok {
		vt := v.([]T)
		o[key] = append(vt, value)
		return
	}

	o[key] = []T{value}
}

// CopyValue copies a value from one option key to another if it exists.
func CopyValue(o Options, fromKey, toKey string) {
	if v, ok := o[fromKey]; ok {
		o[toKey] = v
	}
}

// Get retrieves a value of the specified type from the options.
// If the key does not exist, it returns the provided default value.
// If the value exists but is of a different type, it panics.
func Get[T any](o Options, key string, def T) T {
	v, ok := o[key]
	if !ok {
		return def
	}

	if vt, ok := v.(T); ok {
		return vt
	}

	panic(newTypeMismatchError(key, v, def))
}

// GetInt retrieves an integer value from the options.
// If the key does not exist, it returns the provided default value.
// If the value exists but is of a different integer type,
// it converts it to the desired type.
// If the value is not an integer type, it panics.
func GetInt[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](
	o Options, key string, def T,
) T {
	v, ok := o[key]
	if !ok {
		return def
	}

	switch t := v.(type) {
	case int:
		return T(t)
	case int8:
		return T(t)
	case int16:
		return T(t)
	case int32:
		return T(t)
	case int64:
		return T(t)
	case uint:
		return T(t)
	case uint8:
		return T(t)
	case uint16:
		return T(t)
	case uint32:
		return T(t)
	case uint64:
		return T(t)
	default:
		panic(newTypeMismatchError(key, v, def))
	}
}

// GetFloat retrieves a float value from the options.
// If the key does not exist, it returns the provided default value.
// If the value exists but is of a different float or integer type,
// it converts it to the desired type.
// If the value is not a float or integer type, it panics.
func GetFloat[T ~float32 | ~float64](o Options, key string, def T) T {
	v, ok := o[key]
	if !ok {
		return def
	}

	switch t := v.(type) {
	case int:
		return T(t)
	case int8:
		return T(t)
	case int16:
		return T(t)
	case int32:
		return T(t)
	case int64:
		return T(t)
	case uint:
		return T(t)
	case uint8:
		return T(t)
	case uint16:
		return T(t)
	case uint32:
		return T(t)
	case uint64:
		return T(t)
	case float32:
		return T(t)
	case float64:
		return T(t)
	default:
		panic(newTypeMismatchError(key, v, def))
	}
}

// GetQuality retrieves the quality setting for a given image format.
// It first checks for a general quality setting, then for a format-specific setting,
// and finally falls back to the provided default value if neither is set.
func GetQuality(o Options, format imagetype.Type, def int) int {
	if q := Get(o, keys.Quality, 0); q > 0 {
		return q
	}

	if q := Get(o, keys.FormatQuality+"."+format.String(), 0); q > 0 {
		return q
	}

	return def
}

// GetTime retrieves a [time.Time] value.
// If the key doesn't exist, it returns the zero time.
// If the value exists but is of a different type, it panics.
func GetTime(o Options, key string) time.Time {
	v, ok := o[key]
	if !ok {
		return time.Time{}
	}

	if vt, ok := v.(time.Time); ok {
		return vt
	}

	panic(newTypeMismatchError(key, v, time.Time{}))
}

// GetGravity retrieves a [GravityOptions] value.
// It fills the [GravityOptions] struct with the map values under the given prefix.
// If the gravity type key does not exist,
// it returns a [GravityOptions] with the provided default type.
func GetGravity(o Options, prefix string, defType GravityType) GravityOptions {
	gr := GravityOptions{
		Type: Get(o, prefix+keys.SuffixType, defType),
		X:    GetFloat(o, prefix+keys.SuffixXOffset, 0.0),
		Y:    GetFloat(o, prefix+keys.SuffixYOffset, 0.0),
	}

	return gr
}

// Contains checks if a slice option contains a specific value.
// If the option does not exist, it returns false.
// If the value exists but is of a different type, it panics.
func Contains[T comparable](o Options, key string, value T) bool {
	arr := Get(o, key, []T(nil))
	return slices.Contains(arr, value)
}

// Map returns a copy of the Options as a map[string]any, filtering hidden keys
func (o Options) Map() map[string]any {
	m := make(map[string]any, len(o))

	for k, v := range o {
		if !isHiddenKey(k) {
			m[k] = v
		}
	}

	return m
}

// NestedMap returns Options as a nested map[string]any, filtering hidden keys.
// Each key is split by dots (.) and the resulting keys are used to create a nested structure.
func (o Options) NestedMap() map[string]any {
	nm := make(map[string]any)

	for k, v := range o {
		if isHiddenKey(k) {
			continue
		}
		nestedMapSet(nm, strings.Split(k, "."), v)
	}

	return nm
}

func (o Options) String() string {
	return fmt.Sprintf("%v", o.Map())
}

func (o Options) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.NestedMap())
}

// LogValue returns Options as [slog.Value]
func (o Options) LogValue() slog.Value {
	return toSlogValue(o.NestedMap())
}

// nestedMapSet sets a value in a nested map[string]any structure.
// If the key slice has more than one element, it creates nested maps as needed.
func nestedMapSet(m map[string]any, key []string, value any) {
	if len(key) == 0 {
		return
	}

	if len(key) == 1 {
		m[key[0]] = value
		return
	}

	var mm map[string]any

	if v, ok := m[key[0]]; ok {
		if vm, ok := v.(map[string]any); ok {
			mm = vm
		}
	}

	if mm == nil {
		mm = make(map[string]any)
	}

	nestedMapSet(mm, key[1:], value)

	m[key[0]] = mm
}

func isHiddenKey(key string) bool {
	if len(key) == 0 {
		return false
	}

	// Is key itself hidden?
	if key[0] == '_' {
		return true
	}

	// Is any part of the key hidden?
	return strings.Contains(key, "._")
}

func toSlogValue(v any) slog.Value {
	m, ok := v.(map[string]any)
	if !ok {
		return slog.AnyValue(v)
	}

	attrs := make([]slog.Attr, 0, len(m))

	for k, v := range m {
		attrs = append(attrs, slog.Attr{Key: k, Value: toSlogValue(v)})
	}

	return slog.GroupValue(attrs...)
}
