package fs

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/imgproxy/imgproxy/v3/config"
)

type transport struct {
	fs http.Dir
}

func New() transport {
	return transport{fs: http.Dir(config.LocalFileSystemRoot)}
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	f, err := t.fs.Open(req.URL.Path)

	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("%s is a directory", req.URL.Path)
	}

	header := make(http.Header)

	if config.ETagEnabled {
		etag := BuildEtag(req.URL.Path, fi)
		header.Set("ETag", etag)

		if etag == req.Header.Get("If-None-Match") {
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

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        header,
		ContentLength: fi.Size(),
		Body:          f,
		Close:         true,
		Request:       req,
	}, nil
}

func BuildEtag(path string, fi fs.FileInfo) string {
	tag := fmt.Sprintf("%s__%d__%d", path, fi.Size(), fi.ModTime().UnixNano())
	hash := md5.Sum([]byte(tag))
	return `"` + string(base64.RawURLEncoding.EncodeToString(hash[:])) + `"`
}
