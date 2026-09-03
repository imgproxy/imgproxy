package errorreport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport"
	"github.com/imgproxy/imgproxy/v4/server/meta"
	"github.com/stretchr/testify/require"
)

type fakeReporter struct {
	gotReq  *http.Request
	gotMeta map[string]any
}

func (f *fakeReporter) Report(err errctx.Error, req *http.Request, meta map[string]any) {
	f.gotReq = req
	f.gotMeta = meta
}

func (f *fakeReporter) Close() {}

func TestReportWithNilRequest(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	err := errctx.NewTextError("boom", 0)

	require.NotPanics(t, func() {
		r.Report(context.Background(), err)
	})

	require.Nil(t, fake.gotReq)
	require.Empty(t, fake.gotMeta)
}

func TestReportWithNilRequestAndDocsURL(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	err := errctx.NewTextError("boom", 0, errctx.WithDocsURL("https://example.com/docs"))

	require.NotPanics(t, func() {
		r.Report(context.Background(), err)
	})

	require.Nil(t, fake.gotReq)
	require.Equal(t, map[string]any{"Documentation URL": "https://example.com/docs"}, fake.gotMeta)
}

func TestReportWithRequestAndMetadata(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	ctx := meta.NewContext(req.Context(), req)
	meta.Set(ctx, meta.KeyReqID, "abc123")

	err := errctx.NewTextError("boom", 0, errctx.WithDocsURL("https://example.com/docs"))

	r.Report(ctx, err)

	require.Same(t, req, fake.gotReq)
	require.Equal(t, map[string]any{
		"Request ID":        "abc123",
		"Documentation URL": "https://example.com/docs",
	}, fake.gotMeta)
}

func TestReportPassesThroughAllValues(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	chanValue := make(chan int)

	ctx := meta.NewContext(context.Background(), nil)
	meta.Set(ctx, "String", "value")
	meta.Set(ctx, "Ints", []int{1, 2, 3})
	meta.Set(ctx, "Strings", []string{"a", "b"})
	meta.Set(ctx, "Map", map[string]int{"a": 1, "b": 2})
	meta.Set(ctx, "SliceOfStruct", []struct{ X int }{{X: 1}})
	meta.Set(ctx, "Chan", chanValue)

	err := errctx.NewTextError("boom", 0)

	r.Report(ctx, err)

	require.Equal(t, map[string]any{
		"String":        "value",
		"Ints":          []int{1, 2, 3},
		"Strings":       []string{"a", "b"},
		"Map":           map[string]int{"a": 1, "b": 2},
		"SliceOfStruct": []struct{ X int }{{X: 1}},
		"Chan":          chanValue,
	}, fake.gotMeta)
}

func TestReportWithNilRequestButContextMetadata(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	ctx := meta.NewContext(req.Context(), nil)
	meta.Set(ctx, meta.KeyReqID, "abc123")

	err := errctx.NewTextError("boom", 0)

	r.Report(ctx, err)

	require.Nil(t, fake.gotReq)
	require.Equal(t, map[string]any{"Request ID": "abc123"}, fake.gotMeta)
}
