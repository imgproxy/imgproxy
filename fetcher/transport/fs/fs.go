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

	"github.com/imgproxy/imgproxy/v3/fetcher/transport/common"
	"github.com/imgproxy/imgproxy/v3/fetcher/transport/notmodified"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
)

type transport struct {
	fs             http.Dir
	querySeparator string
}

func New(config *Config, querySeparator string) (transport, error) {
	if err := config.Validate(); err != nil {
		return transport{}, err
	}

	return transport{fs: http.Dir(config.Root), querySeparator: querySeparator}, nil
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	header := make(http.Header)

	_, path, _ := common.GetBucketAndKey(req.URL, t.querySeparator)
	path = "/" + path

	f, err := t.fs.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return respNotFound(req, fmt.Sprintf("%s doesn't exist", path)), nil
		}
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return respNotFound(req, fmt.Sprintf("%s is directory", path)), nil
	}

	statusCode := 200
	size := fi.Size()
	body := io.ReadCloser(f)

	if mimetype := detectContentType(f, fi); len(mimetype) > 0 {
		header.Set(httpheaders.ContentType, mimetype)
	}
	f.Seek(0, io.SeekStart)

	start, end, err := httprange.Parse(req.Header.Get(httpheaders.Range))
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
		header.Set(httpheaders.ContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, fi.Size()))

	default:
		etag := BuildEtag(path, fi)
		header.Set(httpheaders.Etag, etag)

		lastModified := fi.ModTime().Format(http.TimeFormat)
		header.Set(httpheaders.LastModified, lastModified)
	}

	if resp := notmodified.Response(req, header); resp != nil {
		f.Close()
		return resp, nil
	}

	header.Set(httpheaders.AcceptRanges, "bytes")
	header.Set(httpheaders.ContentLength, strconv.Itoa(int(size)))

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
		Header:        http.Header{httpheaders.ContentType: {"text/plain"}},
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
