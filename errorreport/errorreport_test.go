package errorreport_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/imgproxy/imgproxy/v4/errctx"
	"github.com/imgproxy/imgproxy/v4/errorreport"
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
		r.Report(context.Background(), err, nil)
	})

	require.Nil(t, fake.gotReq)
	require.Nil(t, fake.gotMeta)
}

func TestReportWithNilRequestAndDocsURL(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	err := errctx.NewTextError("boom", 0, errctx.WithDocsURL("https://example.com/docs"))

	require.NotPanics(t, func() {
		r.Report(context.Background(), err, nil)
	})

	require.Nil(t, fake.gotReq)
	require.Equal(t, map[string]any{"Documentation URL": "https://example.com/docs"}, fake.gotMeta)
}

func TestReportWithRequestAndMetadata(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	ctx := errorreport.StartRequest(req)
	req = req.WithContext(ctx)
	errorreport.SetMetadata(req, "Request ID", "abc123")

	err := errctx.NewTextError("boom", 0, errctx.WithDocsURL("https://example.com/docs"))

	r.Report(ctx, err, req)

	require.Same(t, req, fake.gotReq)
	require.Equal(t, map[string]any{
		"Request ID":        "abc123",
		"Documentation URL": "https://example.com/docs",
	}, fake.gotMeta)
}

func TestReportWithNilRequestButContextMetadata(t *testing.T) {
	fake := &fakeReporter{}
	r := errorreport.NewWithReporters(fake)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	ctx := errorreport.StartRequest(req)
	req = req.WithContext(ctx)
	errorreport.SetMetadata(req, "Request ID", "abc123")

	err := errctx.NewTextError("boom", 0)

	r.Report(ctx, err, nil)

	require.Nil(t, fake.gotReq)
	require.Equal(t, map[string]any{"Request ID": "abc123"}, fake.gotMeta)
}
