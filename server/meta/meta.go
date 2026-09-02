package meta

import (
	"context"
	"net/http"
	"sync"
)

// Well-known metadata keys, human-readable so they can be used directly as
// error-reporter labels with no translation step.
const (
	KeyReqID             = "Request ID"
	KeyImageURL          = "Source Image URL"
	KeySourceImageOrigin = "Source Image Origin"
	KeyOptions           = "Options"
)

// metaCtxKey is the context key for the Meta value
type metaCtxKey struct{}

// requestCtxKey is the context key for the *http.Request value
type requestCtxKey struct{}

// Meta is generic request-scoped metadata carried via context.
type Meta struct {
	mu     sync.Mutex
	values map[string]any
}

// New creates a new empty Meta.
func New() *Meta {
	return &Meta{
		values: make(map[string]any),
	}
}

// Set stores value under key.
func (m *Meta) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.values[key] = value
}

// Get retrieves a typed value stored under key.
func Get[T any](m *Meta, key string) (v T, ok bool) {
	m.mu.Lock()
	value, exists := m.values[key]
	m.mu.Unlock()

	if !exists {
		return v, false
	}

	v, ok = value.(T)
	return v, ok
}

// Map returns a transformed copy of the metadata; fn returning an empty key drops it.
func (m *Meta) Map(fn func(key string, value any) (string, any)) map[string]any {
	m.mu.Lock()
	defer m.mu.Unlock()

	out := make(map[string]any, len(m.values))

	for k, v := range m.values {
		if nk, nv := fn(k, v); nk != "" {
			out[nk] = nv
		}
	}

	return out
}

// NewContext returns a copy of ctx carrying m and req. req may be nil.
func NewContext(ctx context.Context, m *Meta, req *http.Request) context.Context {
	ctx = context.WithValue(ctx, metaCtxKey{}, m)
	return context.WithValue(ctx, requestCtxKey{}, req)
}

// FromContext returns the Meta attached to ctx, or nil.
func FromContext(ctx context.Context) *Meta {
	m, _ := ctx.Value(metaCtxKey{}).(*Meta)
	return m
}

// RequestFromContext returns the *http.Request attached to ctx, or nil.
func RequestFromContext(ctx context.Context) *http.Request {
	req, _ := ctx.Value(requestCtxKey{}).(*http.Request)
	return req
}

// IsSimpleValue reports whether v is a basic scalar type, or a slice/map of such,
// safe to hand to external systems (loggers, error reporters, etc.) as-is.
func IsSimpleValue(v any) bool {
	switch v.(type) {
	case string, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,

		[]string, []bool,
		[]int, []int8, []int16, []int32, []int64,
		[]uint, []uint8, []uint16, []uint32, []uint64,
		[]float32, []float64,

		map[string]string, map[string]bool,
		map[string]int, map[string]int8, map[string]int16, map[string]int32, map[string]int64,
		map[string]uint, map[string]uint8, map[string]uint16, map[string]uint32, map[string]uint64,
		map[string]float32, map[string]float64:
		return true
	default:
		return false
	}
}
