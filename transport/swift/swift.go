package swift

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/ncw/swift/v2"
)

type transport struct {
	con *swift.Connection
}

func New() (http.RoundTripper, error) {
	c := &swift.Connection{
		UserName:       config.SwiftUsername,
		ApiKey:         config.SwiftAPIKey,
		AuthUrl:        config.SwiftAuthURL,
		AuthVersion:    config.SwiftAuthVersion,
		Domain:         config.SwiftDomain, // v3 auth only
		Tenant:         config.SwiftTenant, // v2 auth only
		Timeout:        time.Duration(config.SwiftTimeoutSeconds) * time.Second,
		ConnectTimeout: time.Duration(config.SwiftConnectTimeoutSeconds) * time.Second,
	}

	ctx := context.Background()

	err := c.Authenticate(ctx)

	if err != nil {
		return nil, fmt.Errorf("swift authentication error: %s", err)
	}

	return transport{con: c}, nil
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	container, path, err := parseObjectURL(req.URL.Path)

	if err != nil {
		return &http.Response{
			Status:     "400 Bad Request",
			StatusCode: 400,
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Close:      true,
			Request:    req,
		}, err
	}

	headers := make(swift.Headers)

	object, headers, err := t.con.ObjectOpen(req.Context(), container, path, false, headers)

	if err != nil {
		return nil, fmt.Errorf("error opening object: %v", err)
	}

	header := make(http.Header)

	if config.ETagEnabled {
		if etag, ok := headers["Etag"]; ok {
			header.Set("ETag", etag)

			if len(etag) > 0 && etag == req.Header.Get("If-None-Match") {
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
	}

	return &http.Response{
		Status:     "200 OK",
		StatusCode: 200,
		Proto:      "HTTP/1.0",
		ProtoMajor: 1,
		ProtoMinor: 0,
		Header:     header,
		Body:       object,
		Close:      true,
		Request:    req,
	}, nil
}

func parseObjectURL(url string) (container string, path string, err error) {
	paths := strings.SplitN(strings.TrimPrefix(url, "/"), "/", 2)

	if len(paths) != 2 {
		return "", "", fmt.Errorf("invalid object url: %s. expecting {container}/{object_path}", url)
	}

	return paths[0], paths[1], nil
}
