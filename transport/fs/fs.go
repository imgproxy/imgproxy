package fs

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/httprange"
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

	statusCode := 200
	size := fi.Size()
	body := io.ReadCloser(f)

	mime := mime.TypeByExtension(filepath.Ext(fi.Name()))
	header.Set("Content-Type", mime)

	start, end, err := httprange.Parse(req.Header.Get("Range"))
	switch {
	case err != nil:
		f.Close()
		return httprange.InvalidHTTPRangeResponse(req), nil

	case end != 0:
		if end < 0 {
			end = size - 1
		}

		f.Seek(start, io.SeekStart)

		statusCode = http.StatusPartialContent
		size = end - start + 1
		body = &fileLimiter{f: f, left: int(size)}
		header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fi.Size()))

	case config.ETagEnabled:
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

	header.Set("Accept-Ranges", "bytes")
	header.Set("Content-Length", strconv.Itoa(int(size)))

	return &http.Response{
		StatusCode:    statusCode,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        header,
		ContentLength: size,
		Body:          body,
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
