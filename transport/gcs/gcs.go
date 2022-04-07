package gcs

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/imgproxy/imgproxy/v3/config"
)

// For tests
var noAuth bool = false

type transport struct {
	client *storage.Client
}

func New() (http.RoundTripper, error) {
	var (
		client *storage.Client
		err    error
	)

	opts := []option.ClientOption{}

	if len(config.GCSKey) > 0 {
		opts = append(opts, option.WithCredentialsJSON([]byte(config.GCSKey)))
	}

	if len(config.GCSEndpoint) > 0 {
		opts = append(opts, option.WithEndpoint(config.GCSEndpoint))
	}

	if noAuth {
		opts = append(opts, option.WithoutAuthentication())
	}

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

	header := make(http.Header)

	if config.ETagEnabled {
		attrs, err := obj.Attrs(req.Context())
		if err != nil {
			return nil, err
		}
		header.Set("ETag", attrs.Etag)

		if etag := req.Header.Get("If-None-Match"); len(etag) > 0 && attrs.Etag == etag {
			return &http.Response{
				StatusCode:    http.StatusNotModified,
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header:        header,
				ContentLength: 0,
				Body:          nil,
				Close:         false,
				Request:       req,
			}, nil
		}
	}

	reader, err := obj.NewReader(req.Context())
	if err != nil {
		return nil, err
	}

	header.Set("Cache-Control", reader.Attrs.CacheControl)

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
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
