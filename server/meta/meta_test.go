package meta_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/imgproxy/imgproxy/v4/server/meta"
	"github.com/stretchr/testify/require"
)

func TestSetGet(t *testing.T) {
	m := meta.New()
	m.Set("key", "value")

	v, ok := meta.Get[string](m, "key")
	require.True(t, ok)
	require.Equal(t, "value", v)
}

func TestGetMissingKey(t *testing.T) {
	m := meta.New()

	_, ok := meta.Get[string](m, "missing")
	require.False(t, ok)
}

func TestGetWrongType(t *testing.T) {
	m := meta.New()
	m.Set("key", 123)

	_, ok := meta.Get[string](m, "key")
	require.False(t, ok)
}

func TestMap(t *testing.T) {
	m := meta.New()
	m.Set("a", 1)
	m.Set("b", 2)

	ctx := meta.NewContext(context.Background(), m, nil)

	out := meta.Map(ctx, func(key string, value any) (string, any) {
		if key == "a" {
			return "", nil
		}
		return key, value
	})

	require.Equal(t, map[string]any{"b": 2}, out)
}

func TestMapNoMeta(t *testing.T) {
	out := meta.Map(context.Background(), func(key string, value any) (string, any) {
		return key, value
	})

	require.Equal(t, map[string]any{}, out)
}

func TestNewContextFromContext(t *testing.T) {
	m := meta.New()
	m.Set("key", "value")
	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	ctx := meta.NewContext(context.Background(), m, req)

	require.Same(t, m, meta.FromContext(ctx))
	require.Same(t, req, meta.RequestFromContext(ctx))
}

func TestNewContextNilRequest(t *testing.T) {
	m := meta.New()

	ctx := meta.NewContext(context.Background(), m, nil)

	require.Same(t, m, meta.FromContext(ctx))
	require.Nil(t, meta.RequestFromContext(ctx))
}

func TestFromContextEmpty(t *testing.T) {
	got := meta.FromContext(context.Background())
	require.Nil(t, got)
}

func TestRequestFromContextEmpty(t *testing.T) {
	got := meta.RequestFromContext(context.Background())
	require.Nil(t, got)
}

func TestConcurrentAccess(t *testing.T) {
	m := meta.New()
	ctx := meta.NewContext(context.Background(), m, nil)

	var wg sync.WaitGroup

	for i := range 100 {
		wg.Add(2)

		go func(i int) {
			defer wg.Done()
			m.Set("key", i)
		}(i)

		go func() {
			defer wg.Done()
			meta.Map(ctx, func(key string, value any) (string, any) {
				return key, value
			})
		}()
	}

	wg.Wait()
}
