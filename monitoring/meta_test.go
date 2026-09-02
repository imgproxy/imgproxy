package monitoring_test

import (
	"context"
	"testing"

	"github.com/imgproxy/imgproxy/v4/monitoring"
	"github.com/imgproxy/imgproxy/v4/options"
	"github.com/imgproxy/imgproxy/v4/server/meta"
	"github.com/stretchr/testify/require"
)

func TestNewMetaFromContextNoMeta(t *testing.T) {
	m := monitoring.NewMetaFromContext(context.Background())
	require.Nil(t, m)
}

func TestNewMetaFromContextTranslatesKnownKeys(t *testing.T) {
	o := options.New()

	cm := meta.New()
	ctx := meta.NewContext(context.Background(), cm, nil)
	meta.Set(ctx, meta.KeyReqID, "abc123") // has no monitoring equivalent
	meta.Set(ctx, meta.KeyImageURL, "http://example.com/image.jpg")
	meta.Set(ctx, meta.KeySourceImageOrigin, "http://example.com")
	meta.Set(ctx, meta.KeyOptions, o)

	m := monitoring.NewMetaFromContext(ctx)

	require.Equal(t, monitoring.Meta{
		monitoring.MetaKey(meta.KeyReqID):             "abc123",
		monitoring.MetaKey(meta.KeyImageURL):          "http://example.com/image.jpg",
		monitoring.MetaKey(meta.KeySourceImageOrigin): "http://example.com",
		monitoring.MetaKey(meta.KeyOptions):           o.Map(),
	}, m)
}

func TestNewMetaFromContextSkipsUnsetKeys(t *testing.T) {
	cm := meta.New()
	ctx := meta.NewContext(context.Background(), cm, nil)
	meta.Set(ctx, meta.KeyImageURL, "http://example.com/image.jpg")

	m := monitoring.NewMetaFromContext(ctx)

	require.Equal(t, monitoring.Meta{
		monitoring.MetaKey(meta.KeyImageURL): "http://example.com/image.jpg",
	}, m)
}

func TestNewMetaFromContextFiltersByKeys(t *testing.T) {
	o := options.New()

	cm := meta.New()
	ctx := meta.NewContext(context.Background(), cm, nil)
	meta.Set(ctx, meta.KeyImageURL, "http://example.com/image.jpg")
	meta.Set(ctx, meta.KeySourceImageOrigin, "http://example.com")
	meta.Set(ctx, meta.KeyOptions, o)

	m := monitoring.NewMetaFromContext(ctx, meta.KeyOptions)

	require.Equal(t, monitoring.Meta{
		monitoring.MetaKey(meta.KeyOptions): o.Map(),
	}, m)
}

func TestNewMetaFromContextIncludesReqIDWhenRequested(t *testing.T) {
	cm := meta.New()
	ctx := meta.NewContext(context.Background(), cm, nil)
	meta.Set(ctx, meta.KeyReqID, "abc123")
	meta.Set(ctx, meta.KeyImageURL, "http://example.com/image.jpg")

	m := monitoring.NewMetaFromContext(ctx, meta.KeyReqID, meta.KeyImageURL)

	require.Equal(t, monitoring.Meta{
		monitoring.MetaKey(meta.KeyReqID):    "abc123",
		monitoring.MetaKey(meta.KeyImageURL): "http://example.com/image.jpg",
	}, m)
}
