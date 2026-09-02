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
	cm.Set(meta.KeyReqID, "abc123") // has no monitoring equivalent
	cm.Set(meta.KeyImageURL, "http://example.com/image.jpg")
	cm.Set(meta.KeySourceImageOrigin, "http://example.com")
	cm.Set(meta.KeyOptions, o)

	ctx := meta.NewContext(context.Background(), cm, nil)

	m := monitoring.NewMetaFromContext(ctx)

	require.Equal(t, monitoring.Meta{
		monitoring.MetaSourceImageURL:    "http://example.com/image.jpg",
		monitoring.MetaSourceImageOrigin: "http://example.com",
		monitoring.MetaOptions:           o.Map(),
	}, m)
}

func TestNewMetaFromContextSkipsUnsetKeys(t *testing.T) {
	cm := meta.New()
	cm.Set(meta.KeyImageURL, "http://example.com/image.jpg")

	ctx := meta.NewContext(context.Background(), cm, nil)

	m := monitoring.NewMetaFromContext(ctx)

	require.Equal(t, monitoring.Meta{
		monitoring.MetaSourceImageURL: "http://example.com/image.jpg",
	}, m)
}

func TestNewMetaFromContextFiltersByKeys(t *testing.T) {
	o := options.New()

	cm := meta.New()
	cm.Set(meta.KeyImageURL, "http://example.com/image.jpg")
	cm.Set(meta.KeySourceImageOrigin, "http://example.com")
	cm.Set(meta.KeyOptions, o)

	ctx := meta.NewContext(context.Background(), cm, nil)

	m := monitoring.NewMetaFromContext(ctx, meta.KeyOptions)

	require.Equal(t, monitoring.Meta{
		monitoring.MetaOptions: o.Map(),
	}, m)
}

func TestNewMetaFromContextFiltersOutReqIDEvenIfRequested(t *testing.T) {
	cm := meta.New()
	cm.Set(meta.KeyReqID, "abc123")
	cm.Set(meta.KeyImageURL, "http://example.com/image.jpg")

	ctx := meta.NewContext(context.Background(), cm, nil)

	m := monitoring.NewMetaFromContext(ctx, meta.KeyReqID, meta.KeyImageURL)

	require.Equal(t, monitoring.Meta{
		monitoring.MetaSourceImageURL: "http://example.com/image.jpg",
	}, m)
}
