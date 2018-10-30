package main

import (
	"context"
	"log"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type gcsTransport struct {
	client *storage.Client
}

func newGCSTransport() http.RoundTripper {
	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON([]byte(conf.GCSKey)))

	if err != nil {
		log.Fatalf("Can't create GCS client: %s", err)
	}

	return gcsTransport{client}
}

func (t gcsTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	bkt := t.client.Bucket(req.URL.Host)
	obj := bkt.Object(strings.TrimPrefix(req.URL.Path, "/"))
	reader, err := obj.NewReader(context.Background())

	if err != nil {
		return nil, err
	}

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        make(http.Header),
		ContentLength: reader.Size(),
		Body:          reader,
		Close:         true,
		Request:       req,
	}, nil
}
