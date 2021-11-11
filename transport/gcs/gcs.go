package gcs

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/imgproxy/imgproxy/v3/config"
	"google.golang.org/api/option"
)

type transport struct {
	client *storage.Client
}

func New() (http.RoundTripper, error) {
	var (
		client *storage.Client
		err    error
	)

	if len(config.GCSKey) > 0 {
		client, err = storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(config.GCSKey)))
	} else {
		client, err = storage.NewClient(context.Background())
	}

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
		attrs, err := obj.Attrs(context.Background())
		if err != nil {
			return nil, err
		}
		header.Set("ETag", attrs.Etag)

		if attrs.Etag == req.Header.Get("If-None-Match") {
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

	reader, err := obj.NewReader(context.Background())
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
