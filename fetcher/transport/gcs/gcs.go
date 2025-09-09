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

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/common"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/notmodified"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

// For tests
var noAuth bool = false

type transport struct {
	client *storage.Client
}

func buildHTTPClient(config *Config, trans *http.Transport, opts ...option.ClientOption) (*http.Client, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	htrans, err := htransport.NewTransport(context.Background(), trans, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "error creating GCS transport")
	}

	return &http.Client{Transport: htrans}, nil
}

func New(config *Config, trans *http.Transport) (http.RoundTripper, error) {
	var client *storage.Client

	opts := []option.ClientOption{
		option.WithScopes(raw.DevstorageReadOnlyScope),
	}

	if len(config.Key) > 0 {
		opts = append(opts, option.WithCredentialsJSON([]byte(config.Key)))
	}

	if len(config.Endpoint) > 0 {
		opts = append(opts, option.WithEndpoint(config.Endpoint))
	}

	if noAuth {
		opts = append(opts, option.WithoutAuthentication())
	}

	httpClient, err := buildHTTPClient(config, trans, opts...)
	if err != nil {
		return nil, err
	}
	opts = append(opts, option.WithHTTPClient(httpClient))

	client, err = storage.NewClient(context.Background(), opts...)

	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("Can't create GCS client"))
	}

	return transport{client}, nil
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	bucket, key, query := common.GetBucketAndKey(req.URL)

	if len(bucket) == 0 || len(key) == 0 {
		body := strings.NewReader("Invalid GCS URL: bucket name or object key is empty")
		return &http.Response{
			StatusCode:    http.StatusNotFound,
			Proto:         "HTTP/1.0",
			ProtoMajor:    1,
			ProtoMinor:    0,
			Header:        http.Header{"Content-Type": {"text/plain"}},
			ContentLength: int64(body.Len()),
			Body:          io.NopCloser(body),
			Close:         false,
			Request:       req,
		}, nil
	}

	bkt := t.client.Bucket(bucket)
	obj := bkt.Object(key)

	if g, err := strconv.ParseInt(query, 10, 64); err == nil && g > 0 {
		obj = obj.Generation(g)
	}

	var (
		reader     *storage.Reader
		statusCode int
		size       int64
	)

	header := make(http.Header)

	if r := req.Header.Get(httpheaders.Range); len(r) != 0 {
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
			header.Set(httpheaders.ContentRange, fmt.Sprintf("bytes %d-%d/%d", reader.Attrs.StartOffset, end, reader.Attrs.Size))
		}
	}

	// We haven't initialize reader yet, this means that we need non-ranged reader
	if reader == nil {
		attrs, aerr := obj.Attrs(req.Context())
		if aerr != nil {
			return handleError(req, aerr)
		}
		header.Set(httpheaders.Etag, attrs.Etag)
		header.Set(httpheaders.LastModified, attrs.Updated.Format(http.TimeFormat))

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

	header.Set(httpheaders.AcceptRanges, "bytes")
	header.Set(httpheaders.ContentLength, strconv.Itoa(int(size)))
	header.Set(httpheaders.ContentType, reader.Attrs.ContentType)
	header.Set(httpheaders.CacheControl, reader.Attrs.CacheControl)

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
		Header:        http.Header{httpheaders.ContentType: {"text/plain"}},
		ContentLength: int64(len(err.Error())),
		Body:          io.NopCloser(strings.NewReader(err.Error())),
		Close:         false,
		Request:       req,
	}, nil
}
