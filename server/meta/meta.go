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

// meta is generic request-scoped metadata carried via context.
type meta struct {
	mu     sync.Mutex
	values map[string]any
}

// New creates a new empty Meta.
func New() *meta {
	return &meta{
		values: make(map[string]any),
	}
}

// Set stores value under key in the Meta attached to ctx. No-op if ctx carries no *Meta.
func Set(ctx context.Context, key string, value any) {
	m := FromContext(ctx)
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.values[key] = value
}

// Get retrieves a typed value stored under key.
func Get[T any](m *meta, key string) (v T, ok bool) {
	m.mu.Lock()
	value, exists := m.values[key]
	m.mu.Unlock()

	if !exists {
		return v, false
	}

	v, ok = value.(T)
	return v, ok
}

// Map returns a transformed copy of the metadata attached to ctx; fn returning an
// empty key drops it. Returns an empty map if ctx carries no *Meta.
func Map(ctx context.Context, fn func(key string, value any) (string, any)) map[string]any {
	m := FromContext(ctx)
	if m == nil {
		return map[string]any{}
	}

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
func NewContext(ctx context.Context, m *meta, req *http.Request) context.Context {
	ctx = context.WithValue(ctx, metaCtxKey{}, m)
	return context.WithValue(ctx, requestCtxKey{}, req)
}

// FromContext returns the Meta attached to ctx, or nil.
func FromContext(ctx context.Context) *meta {
	m, _ := ctx.Value(metaCtxKey{}).(*meta)
	return m
}

// RequestFromContext returns the *http.Request attached to ctx, or nil.
func RequestFromContext(ctx context.Context) *http.Request {
	req, _ := ctx.Value(requestCtxKey{}).(*http.Request)
	return req
}
