package fs

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"
)

type transport struct {
	fs http.Dir
}

func New() transport {
	return transport{fs: http.Dir(config.LocalFileSystemRoot)}
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	header := make(http.Header)

	f, err := t.fs.Open(req.URL.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return respNotFound(req, fmt.Sprintf("%s doesn't exist", req.URL.Path)), nil
		}
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return respNotFound(req, fmt.Sprintf("%s is directory", req.URL.Path)), nil
	}

	if config.ETagEnabled {
		etag := BuildEtag(req.URL.Path, fi)
		header.Set("ETag", etag)

		if etag == req.Header.Get("If-None-Match") {
			f.Close()

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

func respNotFound(req *http.Request, msg string) *http.Response {
	return &http.Response{
		StatusCode:    http.StatusNotFound,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        make(http.Header),
		ContentLength: int64(len(msg)),
		Body:          io.NopCloser(strings.NewReader(msg)),
		Close:         false,
		Request:       req,
	}
}
