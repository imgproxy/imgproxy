package notmodified

import (
	"net/http"
	"time"

	"github.com/imgproxy/imgproxy/v3/config"
)

func Response(req *http.Request, header http.Header) *http.Response {
	if config.ETagEnabled {
		etag := header.Get("ETag")
		ifNoneMatch := req.Header.Get("If-None-Match")

		if len(ifNoneMatch) > 0 && ifNoneMatch == etag {
			return response(req, header)
		}
	}
	if config.LastModifiedEnabled {
		lastModifiedRaw := header.Get("Last-Modified")
		if len(lastModifiedRaw) == 0 {
			return nil
		}
		ifModifiedSinceRaw := req.Header.Get("If-Modified-Since")
		if len(ifModifiedSinceRaw) == 0 {
			return nil
		}
		lastModified, err := time.Parse(http.TimeFormat, lastModifiedRaw)
		if err != nil {
			return nil
		}
		ifModifiedSince, err := time.Parse(http.TimeFormat, ifModifiedSinceRaw)
		if err != nil {
			return nil
		}
		if !ifModifiedSince.Before(lastModified) {
			return response(req, header)
		}
	}

	return nil
}

func response(req *http.Request, header http.Header) *http.Response {
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
	}
}
