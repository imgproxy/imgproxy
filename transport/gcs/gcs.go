package gcs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
	raw "google.golang.org/api/storage/v1"
	htransport "google.golang.org/api/transport/http"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/httprange"
	defaultTransport "github.com/imgproxy/imgproxy/v3/transport"
	"github.com/imgproxy/imgproxy/v3/transport/notmodified"
)

// For tests
var noAuth bool = false

type transport struct {
	client *storage.Client
}

func buildHTTPClient(opts ...option.ClientOption) (*http.Client, error) {
	trans, err := defaultTransport.New(false)
	if err != nil {
		return nil, err
	}

	htrans, err := htransport.NewTransport(context.Background(), trans, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error creating GCS transport")
	}

	return &http.Client{Transport: htrans}, nil
}

func New() (http.RoundTripper, error) {
	var client *storage.Client

	opts := []option.ClientOption{
		option.WithScopes(raw.DevstorageReadOnlyScope),
		option.WithUserAgent(config.UserAgent),
	}

	if len(config.GCSKey) > 0 {
		opts = append(opts, option.WithCredentialsJSON([]byte(config.GCSKey)))
	}

	if len(config.GCSEndpoint) > 0 {
		opts = append(opts, option.WithEndpoint(config.GCSEndpoint))
	}

	if noAuth {
		opts = append(opts, option.WithoutAuthentication())
	}

	httpClient, err := buildHTTPClient(opts...)
	if err != nil {
		return nil, err
	}
	opts = append(opts, option.WithHTTPClient(httpClient))

	client, err = storage.NewClient(context.Background(), opts...)

	if err != nil {
		return nil, fmt.Errorf("Can't create GCS client: %s", err)
	}

	return transport{client}, nil
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	bkt := t.client.Bucket(req.URL.Host)
	obj := bkt.Object(strings.TrimPrefix(req.URL.Path, "/"))

	if g, err := strconv.ParseInt(req.URL.RawQuery, 10, 64); err == nil && g > 0 {
		obj = obj.Generation(g)
	}

	var (
		reader     *storage.Reader
		statusCode int
		size       int64
	)

	header := make(http.Header)

	if r := req.Header.Get("Range"); len(r) != 0 {
		start, end, err := httprange.Parse(r)
		if err != nil {
			return httprange.InvalidHTTPRangeResponse(req), nil
		}

		if end != 0 {
			length := end - start + 1
			if end < 0 {
				length = -1
			}

			reader, err = obj.NewRangeReader(req.Context(), start, length)
			if err != nil {
				return nil, err
			}

			if end < 0 || end >= reader.Attrs.Size {
				end = reader.Attrs.Size - 1
			}

			size = end - reader.Attrs.StartOffset + 1

			statusCode = http.StatusPartialContent
			header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", reader.Attrs.StartOffset, end, reader.Attrs.Size))
		}
	}

	// We haven't initialize reader yet, this means that we need non-ranged reader
	if reader == nil {
		if config.ETagEnabled || config.LastModifiedEnabled {
			attrs, err := obj.Attrs(req.Context())
			if err != nil {
				return handleError(req, err)
			}
			if config.ETagEnabled {
				header.Set("ETag", attrs.Etag)
			}
			if config.LastModifiedEnabled {
				header.Set("Last-Modified", attrs.Updated.Format(http.TimeFormat))
			}
		}

		if resp := notmodified.Response(req, header); resp != nil {
			return resp, nil
		}

		var err error
		reader, err = obj.NewReader(req.Context())
		if err != nil {
			return handleError(req, err)
		}

		statusCode = 200
		size = reader.Attrs.Size
	}

	header.Set("Accept-Ranges", "bytes")
	header.Set("Content-Length", strconv.Itoa(int(size)))
	header.Set("Content-Type", reader.Attrs.ContentType)
	header.Set("Cache-Control", reader.Attrs.CacheControl)

	return &http.Response{
		StatusCode:    statusCode,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        header,
		ContentLength: reader.Attrs.Size,
		Body:          reader,
		Close:         true,
		Request:       req,
	}, nil
}

func handleError(req *http.Request, err error) (*http.Response, error) {
	if err != storage.ErrBucketNotExist && err != storage.ErrObjectNotExist {
		return nil, err
	}

	return &http.Response{
		StatusCode:    http.StatusNotFound,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        make(http.Header),
		ContentLength: int64(len(err.Error())),
		Body:          io.NopCloser(strings.NewReader(err.Error())),
		Close:         false,
		Request:       req,
	}, nil
}
