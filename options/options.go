package options

import (
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Options is an interface for storing and retrieving dynamic option values.
//
// Copies of Options are shallow, meaning the underlying map is shared.
type Options struct {
	m map[string]any

	main  *Options // Pointer to the main Options if this is a child
	child *Options // Pointer to the child Options
}

// New creates a new Options map
func New() *Options {
	return &Options{
		m: make(map[string]any),
	}
}

// Main returns the main Options if this is a child Options.
// If this is the main Options, it returns itself.
func (o *Options) Main() *Options {
	if o.main == nil {
		return o
	}

	return o.main
}

// Child returns the child Options if any.
func (o *Options) Child() *Options {
	return o.child
}

// Descendants returns an iterator over the child Options if any.
func (o *Options) Descendants() iter.Seq[*Options] {
	return func(yield func(*Options) bool) {
		for c := o.child; c != nil; c = c.child {
			if !yield(c) {
				return
			}
		}
	}
}

// HasChild checks if the Options has a child Options.
func (o *Options) HasChild() bool {
	return o.child != nil
}

// AddChild creates a new child Options that inherits from the current Options.
// If the current Options already has a child, it returns the existing child.
func (o *Options) AddChild() *Options {
	if o.child != nil {
		return o.child
	}

	child := New()
	child.main = o.Main()
	o.child = child

	return child
}

// Depth returns the depth of the Options in the hierarchy.
// The main Options has a depth of 0, its child has a depth of 1, and so on.
func (o *Options) Depth() int {
	depth := 0

	for p := o.main; p != nil && p != o; p = p.child {
		depth++
	}

	return depth
}

// Get retrieves a value of the specified type from the options.
// If the key does not exist, it returns the provided default value.
// If the value exists but is of a different type, it panics.
func Get[T any](o *Options, key string, def T) T {
	v, ok := o.m[key]
	if !ok {
		return def
	}

	if vt, ok := v.(T); ok {
		return vt
	}

	panic(newTypeMismatchError(key, v, def))
}

// AppendToSlice appends a value to a slice option.
// If the option does not exist, it creates a new slice with the value.
func AppendToSlice[T any](o *Options, key string, value ...T) {
	if v, ok := o.m[key]; ok {
		vt := v.([]T)
		o.m[key] = append(vt, value...)
		return
	}

	o.m[key] = append([]T(nil), value...)
}

// SliceContains checks if a slice option contains a specific value.
// If the option does not exist, it returns false.
// If the value exists but is of a different type, it panics.
func SliceContains[T comparable](o *Options, key string, value T) bool {
	arr := Get(o, key, []T(nil))
	return slices.Contains(arr, value)
}

// Set sets a value for a specific option key.
func (o *Options) Set(key string, value any) {
	o.m[key] = value
}

// Propagate propagates a value under the given key to the child Options if any.
func (o *Options) Propagate(key string) {
	if o.child == nil {
		return
	}

	if v, ok := o.m[key]; ok {
		for c := range o.Descendants() {
			c.m[key] = v
		}
	}
}

// Delete removes an option by its key.
func (o *Options) Delete(key string) {
	delete(o.m, key)
}

// DeleteFromChildren removes an option by its key from the child Options if any.
func (o *Options) DeleteFromChildren(key string) {
	if o.child == nil {
		return
	}

	for c := range o.Descendants() {
		delete(c.m, key)
	}
}

// CopyValue copies a value from one option key to another.
func (o *Options) CopyValue(fromKey, toKey string) {
	if v, ok := o.m[fromKey]; ok {
		o.m[toKey] = v
	}
}

// Has checks if an option key exists.
func (o *Options) Has(key string) bool {
	_, ok := o.m[key]
	return ok
}

// GetInt retrieves an int value from the options.
// If the key does not exist, GetInt returns the provided default value.
// If the key exists but the value is of a different integer type,
// GetInt converts it to int.
// If the key exists but the value is not an integer type, GetInt panics.
func (o *Options) GetInt(key string, def int) int {
	v, ok := o.m[key]
	if !ok {
		return def
	}

	switch t := v.(type) {
	case int:
		return t
	case int8:
		return int(t)
	case int16:
		return int(t)
	case int32:
		return int(t)
	case int64:
		return int(t)
	case uint:
		return int(t)
	case uint8:
		return int(t)
	case uint16:
		return int(t)
	case uint32:
		return int(t)
	case uint64:
		return int(t)
	default:
		panic(newTypeMismatchError(key, v, def))
	}
}

// GetFloat retrieves a float64 value from the options.
// If the key does not exist, GetFloat returns the provided default value.
// If the key value exists but the value is of a different float or integer type,
// GetFloat converts it to float64.
// If the key exists but the value is not a float or integer type, GetFloat panics.
func (o *Options) GetFloat(key string, def float64) float64 {
	v, ok := o.m[key]
	if !ok {
		return def
	}

	switch t := v.(type) {
	case int:
		return float64(t)
	case int8:
		return float64(t)
	case int16:
		return float64(t)
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case uint:
		return float64(t)
	case uint8:
		return float64(t)
	case uint16:
		return float64(t)
	case uint32:
		return float64(t)
	case uint64:
		return float64(t)
	case float32:
		return float64(t)
	case float64:
		return t
	default:
		panic(newTypeMismatchError(key, v, def))
	}
}

// GetString retrieves a string value.
// If the key doesn't exist, it returns the provided default value.
// If the value exists but is of a different type, it panics.
func (o *Options) GetString(key string, def string) string {
	return Get(o, key, def)
}

// GetBool retrieves a bool value.
// If the key doesn't exist, it returns the provided default value.
// If the value exists but is of a different type, it panics.
func (o *Options) GetBool(key string, def bool) bool {
	return Get(o, key, def)
}

// GetTime retrieves a [time.Time] value.
// If the key doesn't exist, it returns the zero time.
// If the value exists but is of a different type, it panics.
func (o *Options) GetTime(key string) time.Time {
	return Get(o, key, time.Time{})
}

// Map returns a copy of the Options as a map[string]any
// If the Options has a child, it combines the main and child maps,
// prepending each key with the options depth
// (e.g., "0.key" for main options, "1.key" for child options, "2.key" for grandchild options, etc.)
func (o *Options) Map() map[string]any {
	if o.child == nil {
		return maps.Clone(o.m)
	}

	totalEntries := len(o.m)

	for c := range o.Descendants() {
		totalEntries += len(c.m)
	}

	result := make(map[string]any, totalEntries)

	for k, v := range o.m {
		result["0."+k] = v
	}

	depth := 1
	for c := range o.Descendants() {
		for k, v := range c.m {
			result[strconv.Itoa(depth)+"."+k] = v
		}
		depth++
	}

	return result
}

// NestedMap returns Options as a nested map[string]any.
// Each key is split by dots (.) and the resulting keys are used to create a nested structure.
// If the Options has a child, it puts the main and child maps under "0", "1", "2", etc. keys
// representing the options depth
// (e.g., "0" for main options, "1" for child options, "2" for grandchild options, etc.)
func (o *Options) NestedMap() map[string]any {
	if o.child == nil {
		return o.nestedMap()
	}

	totalMaps := 1
	for child := o.child; child != nil; child = child.child {
		totalMaps++
	}

	result := make(map[string]any, totalMaps)

	result["0"] = o.nestedMap()

	depth := 1
	for c := range o.Descendants() {
		result[strconv.Itoa(depth)] = c.nestedMap()
		depth++
	}

	return result
}

func (o *Options) nestedMap() map[string]any {
	nm := make(map[string]any)

	for k, v := range o.m {
		nestedMapSet(nm, k, v)
	}

	return nm
}

// String returns Options as a string representation of the map.
func (o *Options) String() string {
	return fmt.Sprintf("%v", o.Map())
}

// MarshalJSON returns Options as a JSON byte slice.
func (o *Options) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.NestedMap())
}

// LogValue returns Options as [slog.Value]
func (o *Options) LogValue() slog.Value {
	return toSlogValue(o.NestedMap())
}

// nestedMapSet sets a value in a nested map[string]any structure.
// If the key has more than one element, it creates nested maps as needed.
func nestedMapSet(m map[string]any, key string, value any) {
	key, rest, isGroup := strings.Cut(key, ".")

	if !isGroup {
		m[key] = value
		return
	}

	mm, ok := m[key].(map[string]any)
	if !ok {
		mm = make(map[string]any)
	}

	nestedMapSet(mm, rest, value)

	m[key] = mm
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

	// Sort attributes by key to have a consistent order
	slices.SortFunc(attrs, func(a, b slog.Attr) int {
		return strings.Compare(a.Key, b.Key)
	})

	return slog.GroupValue(attrs...)
}
