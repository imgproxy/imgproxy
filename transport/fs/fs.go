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
	"github.com/imgproxy/imgproxy/v3/transport/notmodified"
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

	if mimetype := detectContentType(f, fi); len(mimetype) > 0 {
		header.Set("Content-Type", mimetype)
	}
	f.Seek(0, io.SeekStart)

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

	default:
		if config.ETagEnabled {
			etag := BuildEtag(req.URL.Path, fi)
			header.Set("ETag", etag)
		}

		if config.LastModifiedEnabled {
			lastModified := fi.ModTime().Format(http.TimeFormat)
			header.Set("Last-Modified", lastModified)
		}
	}

	if resp := notmodified.Response(req, header); resp != nil {
		f.Close()
		return resp, nil
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

func detectContentType(f http.File, fi fs.FileInfo) string {
	var (
		tmp      [512]byte
		mimetype string
	)

	if n, err := io.ReadFull(f, tmp[:]); err == nil {
		mimetype = http.DetectContentType(tmp[:n])
	}

	if len(mimetype) == 0 || strings.HasPrefix(mimetype, "text/plain") || strings.HasPrefix(mimetype, "application/octet-stream") {
		if m := mime.TypeByExtension(filepath.Ext(fi.Name())); len(m) > 0 {
			mimetype = m
		}
	}

	return mimetype
}
