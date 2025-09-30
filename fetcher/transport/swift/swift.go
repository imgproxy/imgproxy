package swift

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/ncw/swift/v2"

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/common"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/notmodified"
	"github.com/imgproxy/imgproxy/v3/ierrors"
)

type transport struct {
	con            *swift.Connection
	querySeparator string
}

func New(config *Config, trans *http.Transport, querySeparator string) (http.RoundTripper, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	c := &swift.Connection{
		UserName:       config.Username,
		ApiKey:         config.APIKey,
		AuthUrl:        config.AuthURL,
		AuthVersion:    config.AuthVersion,
		Domain:         config.Domain, // v3 auth only
		Tenant:         config.Tenant, // v2 auth only
		Timeout:        config.Timeout,
		ConnectTimeout: config.ConnectTimeout,
		Transport:      trans,
	}

	ctx := context.Background()

	err := c.Authenticate(ctx)

	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("swift authentication error"))
	}

	return transport{con: c, querySeparator: querySeparator}, nil
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	container, objectName, _ := common.GetBucketAndKey(req.URL, t.querySeparator)

	if len(container) == 0 || len(objectName) == 0 {
		body := strings.NewReader("Invalid Swift URL: container name or object name is empty")
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

	reqHeaders := make(swift.Headers)
	if r := req.Header.Get("Range"); len(r) > 0 {
		reqHeaders["Range"] = r
	}

	object, objectHeaders, err := t.con.ObjectOpen(req.Context(), container, objectName, false, reqHeaders)

	header := make(http.Header)

	if err != nil {
		if errors.Is(err, swift.ObjectNotFound) || errors.Is(err, swift.ContainerNotFound) {
			return &http.Response{
				StatusCode:    http.StatusNotFound,
				Proto:         "HTTP/1.0",
				ProtoMajor:    1,
				ProtoMinor:    0,
				Header:        http.Header{"Content-Type": {"text/plain"}},
				ContentLength: int64(len(err.Error())),
				Body:          io.NopCloser(strings.NewReader(err.Error())),
				Close:         false,
				Request:       req,
			}, nil
		}

		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("error opening object"))
	}

	if etag, ok := objectHeaders["Etag"]; ok {
		header.Set("ETag", etag)
	}

	if lastModified, ok := objectHeaders["Last-Modified"]; ok {
		header.Set("Last-Modified", lastModified)
	}

	if resp := notmodified.Response(req, header); resp != nil {
		object.Close()
		return resp, nil
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
