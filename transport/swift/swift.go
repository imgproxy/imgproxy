package swift

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ncw/swift/v2"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	defaultTransport "github.com/imgproxy/imgproxy/v3/transport"
	"github.com/imgproxy/imgproxy/v3/transport/common"
	"github.com/imgproxy/imgproxy/v3/transport/notmodified"
)

type transport struct {
	con *swift.Connection
}

func New() (http.RoundTripper, error) {
	trans, err := defaultTransport.New(false)
	if err != nil {
		return nil, err
	}

	c := &swift.Connection{
		UserName:       config.SwiftUsername,
		ApiKey:         config.SwiftAPIKey,
		AuthUrl:        config.SwiftAuthURL,
		AuthVersion:    config.SwiftAuthVersion,
		Domain:         config.SwiftDomain, // v3 auth only
		Tenant:         config.SwiftTenant, // v2 auth only
		Timeout:        time.Duration(config.SwiftTimeoutSeconds) * time.Second,
		ConnectTimeout: time.Duration(config.SwiftConnectTimeoutSeconds) * time.Second,
		Transport:      trans,
	}

	ctx := context.Background()

	err = c.Authenticate(ctx)

	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("swift authentication error"))
	}

	return transport{con: c}, nil
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	container, objectName, _ := common.GetBucketAndKey(req.URL)

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

	if config.ETagEnabled {
		if etag, ok := objectHeaders["Etag"]; ok {
			header.Set("ETag", etag)
		}
	}

	if config.LastModifiedEnabled {
		if lastModified, ok := objectHeaders["Last-Modified"]; ok {
			header.Set("Last-Modified", lastModified)
		}
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
