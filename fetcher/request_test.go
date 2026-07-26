package fetcher_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/imgproxy/imgproxy/v4/fetcher"
	"github.com/imgproxy/imgproxy/v4/httpheaders"
)

func TestFetchGzipResponseMetadata(t *testing.T) {
	source := []byte("gzip response body")
	compressed := new(bytes.Buffer)

	enc := gzip.NewWriter(compressed)
	_, err := enc.Write(source)
	require.NoError(t, err)
	require.NoError(t, enc.Close())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set(httpheaders.ContentEncoding, "gzip")
		w.Header().Set(httpheaders.ContentLength, strconv.Itoa(compressed.Len()))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(compressed.Bytes())
	}))
	defer server.Close()

	config := fetcher.NewDefaultConfig()
	config.Transport.HTTP.AllowLoopbackSourceAddresses = true

	f, err := fetcher.New(&config)
	require.NoError(t, err)

	req, err := f.BuildRequest(context.Background(), server.URL, nil, nil)
	require.NoError(t, err)
	defer req.Cancel()

	res, err := req.Fetch()
	require.NoError(t, err)
	defer res.Body.Close()

	require.Equal(t, int64(-1), res.ContentLength)
	require.Empty(t, res.Header.Get(httpheaders.ContentEncoding))
	require.Empty(t, res.Header.Get(httpheaders.ContentLength))
	require.True(t, res.Uncompressed)

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, source, body)
}
