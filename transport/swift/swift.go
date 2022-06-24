package swift

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ncw/swift/v2"

	"github.com/imgproxy/imgproxy/v3/config"
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
	// Users should have converted the object storage URL in the format of swift://{container}/{object}
	container := req.URL.Host
	objectName := strings.TrimPrefix(req.URL.Path, "/")

	object, objectHeaders, err := t.con.ObjectOpen(req.Context(), container, objectName, false, make(swift.Headers))

	header := make(http.Header)

	if err != nil {
		if errors.Is(err, swift.ObjectNotFound) || errors.Is(err, swift.ContainerNotFound) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Proto:      "HTTP/1.0",
				ProtoMajor: 1,
				ProtoMinor: 0,
				Header:     header,
				Body:       io.NopCloser(strings.NewReader(err.Error())),
				Close:      false,
				Request:    req,
			}, nil
		}

		return nil, fmt.Errorf("error opening object: %v", err)
	}

	if config.ETagEnabled {
		if etag, ok := objectHeaders["Etag"]; ok {
			header.Set("ETag", etag)

			if len(etag) > 0 && etag == req.Header.Get("If-None-Match") {
				object.Close()

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

	for k, v := range objectHeaders {
		header.Set(k, v)
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
