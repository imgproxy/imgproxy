package fs

import (
	"context"
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

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
	"github.com/imgproxy/imgproxy/v3/storage/common"
	"github.com/imgproxy/imgproxy/v3/storage/response"
)

// Storage represents fs file storage
type Storage struct {
	fs             http.Dir
	querySeparator string
}

// New creates a new Storage instance.
func New(config *Config, qsSeparator string) (*Storage, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Storage{fs: http.Dir(config.Root), querySeparator: qsSeparator}, nil
}

// GetObject retrieves an object from file system.
func (s *Storage) GetObject(
	ctx context.Context,
	reqHeader http.Header,
	_, name, _ string,
) (*response.Object, error) {
	// If either container or object name is empty, return 404
	if len(name) == 0 {
		return response.NewNotFound(
			"invalid FS Storage URL: object name is empty",
		), nil
	}

	name = "/" + name

	// check that file exists
	f, err := s.fs.Open(name)
	if err != nil {
		if os.IsNotExist(err) {
			return response.NewNotFound(fmt.Sprintf("%s doesn't exist", name)), nil
		}

		return nil, err
	}

	// check that file is not a directory
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return response.NewNotFound(fmt.Sprintf("%s is directory", name)), nil
	}

	// file basic properties
	size := fi.Size()
	body := io.ReadCloser(f)

	// result headers
	header := make(http.Header)

	// set default headers
	header.Set(httpheaders.AcceptRanges, "bytes")

	// try to detect content type from magic bytes or extension
	if mimetype := detectContentType(f, fi); len(mimetype) > 0 {
		header.Set(httpheaders.ContentType, mimetype)
	}

	// calculate Etag and Last-Modified date
	etag := buildEtag(name, fi)
	lastModified := fi.ModTime().Format(http.TimeFormat)

	// try requested range
	start, end, err := httprange.Parse(header.Get(httpheaders.Range))
	switch {
	case err != nil:
		f.Close()
		return response.NewInvalidRange(), nil

	// Range requested: partial content should be returned
	case end != 0:
		if end < 0 {
			end = size - 1
		}

		f.Seek(start, io.SeekStart)

		size = end - start + 1
		body = &fileLimiter{f: f, left: int(size)}
		header.Set(httpheaders.ContentRange, fmt.Sprintf("bytes %d-%d/%d", start, end, fi.Size()))

		return response.NewPartialContent(header, body), nil

	// Full object requested
	default:
		header.Set(httpheaders.Etag, etag)
		header.Set(httpheaders.LastModified, lastModified)
	}

	// Either size of a partial or the total
	header.Set(httpheaders.ContentLength, strconv.Itoa(int(size)))

	// In case file was not modified, let's not return reader
	if common.IsNotModified(reqHeader, header) {
		f.Close()
		return response.NewNotModified(header), nil
	}

	return response.NewOK(header, body), nil
}

func buildEtag(path string, fi fs.FileInfo) string {
	tag := fmt.Sprintf("%s__%d__%d", path, fi.Size(), fi.ModTime().UnixNano())
	hash := md5.Sum([]byte(tag))
	return `"` + string(base64.RawURLEncoding.EncodeToString(hash[:])) + `"`
}

// detectContentType detects the content type of a file by mime or extension
func detectContentType(f http.File, fi fs.FileInfo) string {
	var (
		tmp      [512]byte
		mimetype string
	)

	if n, err := io.ReadFull(f, tmp[:]); err == nil || err == io.ErrUnexpectedEOF {
		mimetype = http.DetectContentType(tmp[:n])
	}

	f.Seek(0, io.SeekStart) // rewind file position

	if len(mimetype) == 0 || strings.HasPrefix(mimetype, "text/plain") || strings.HasPrefix(mimetype, "application/octet-stream") {
		if m := mime.TypeByExtension(filepath.Ext(fi.Name())); len(m) > 0 {
			mimetype = m
		}
	}

	return mimetype
}
