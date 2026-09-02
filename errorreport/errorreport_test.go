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

	m := meta.New()
	m.Set(meta.KeyReqID, "abc123")
	ctx := meta.NewContext(req.Context(), m, req)

	err := errctx.NewTextError("boom", 0, errctx.WithDocsURL("https://example.com/docs"))

	r.Report(ctx, err)

	require.Same(t, req, fake.gotReq)
	require.Equal(t, map[string]any{
		"Request ID":        "abc123",
		"Documentation URL": "https://example.com/docs",
	}, fake.gotMeta)
}

func TestReportFiltersNonSimpleValues(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	m := meta.New()
	m.Set("String", "value")
	m.Set("Ints", []int{1, 2, 3})
	m.Set("Strings", []string{"a", "b"})
	m.Set("Map", map[string]int{"a": 1, "b": 2})
	m.Set("SliceOfStruct", []struct{ X int }{{X: 1}})
	m.Set("Chan", make(chan int))

	ctx := meta.NewContext(context.Background(), m, nil)

	err := errctx.NewTextError("boom", 0)

	r.Report(ctx, err)

	require.Equal(t, map[string]any{
		"String":  "value",
		"Ints":    []int{1, 2, 3},
		"Strings": []string{"a", "b"},
		"Map":     map[string]int{"a": 1, "b": 2},
	}, fake.gotMeta)
}

func TestReportWithNilRequestButContextMetadata(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	m := meta.New()
	m.Set(meta.KeyReqID, "abc123")
	ctx := meta.NewContext(req.Context(), m, nil)

	err := errctx.NewTextError("boom", 0)

	r.Report(ctx, err)

	require.Nil(t, fake.gotReq)
	require.Equal(t, map[string]any{"Request ID": "abc123"}, fake.gotMeta)
}
