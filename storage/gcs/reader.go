package gcs

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	gcs "cloud.google.com/go/storage"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/httprange"
	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/common"
	"github.com/pkg/errors"
)

// GetObject retrieves an object from Azure cloud
func (s *Storage) GetObject(
	ctx context.Context,
	reqHeader http.Header,
	bucket, key, query string,
) (*storage.ObjectReader, error) {
	// If either bucket or object key is empty, return 404
	if len(bucket) == 0 || len(key) == 0 {
		return storage.NewObjectNotFound(
			"invalid GCS Storage URL: bucket name or object key are empty",
		), nil
	}

	// Check if access to the bucket is allowed
	if !common.IsBucketAllowed(bucket, s.config.AllowedBuckets, s.config.DeniedBuckets) {
		return nil, fmt.Errorf("access to the GCS bucket %s is denied", bucket)
	}

	var (
		reader *gcs.Reader
		size   int64
	)

	bkt := s.client.Bucket(bucket)
	obj := bkt.Object(key)

	if g, err := strconv.ParseInt(query, 10, 64); err == nil && g > 0 {
		obj = obj.Generation(g)
	}

	header := make(http.Header)

	// Try respond with partial: if that was a partial request,
	// we either return error or Object
	if r, err := s.tryRespondWithPartial(ctx, obj, reqHeader, header); r != nil || err != nil {
		return r, err
	}

	var err error
	reader, err = obj.NewReader(ctx)
	if err != nil {
		return handleError(err)
	}

	// Generate artificial ETag from CRC32 and LastModified
	var etag [12]byte
	binary.LittleEndian.PutUint32(etag[:4], uint32(reader.Attrs.CRC32C))
	binary.LittleEndian.PutUint64(etag[4:], uint64(reader.Attrs.LastModified.UnixNano()))

	header.Set(httpheaders.Etag, hex.EncodeToString(etag[:]))
	header.Set(httpheaders.LastModified, reader.Attrs.LastModified.Format(http.TimeFormat))

	if common.IsNotModified(reqHeader, header) {
		reader.Close()
		return storage.NewObjectNotModified(header), nil
	}

	size = reader.Attrs.Size
	setHeadersFromReader(header, reader, size)

	return storage.NewObjectOK(header, reader), nil
}

// tryRespondWithPartial tries to respond with a partial object
// if the Range header is set.
func (s *Storage) tryRespondWithPartial(
	ctx context.Context,
	obj *gcs.ObjectHandle,
	reqHeader http.Header,
	header http.Header,
) (*storage.ObjectReader, error) {
	r := reqHeader.Get(httpheaders.Range)
	if len(r) == 0 {
		return nil, nil
	}

	start, end, err := httprange.Parse(r)
	if err != nil {
		return storage.NewObjectInvalidRange(), nil
	}

	if end == 0 {
		return nil, nil
	}

	length := end - start + 1
	if end < 0 {
		length = -1
	}

	reader, err := obj.NewRangeReader(ctx, start, length)
	if err != nil {
		return nil, err
	}

	if end < 0 || end >= reader.Attrs.Size {
		end = reader.Attrs.Size - 1
	}

	size := end - reader.Attrs.StartOffset + 1

	header.Set(httpheaders.ContentRange, fmt.Sprintf("bytes %d-%d/%d", reader.Attrs.StartOffset, end, reader.Attrs.Size))
	setHeadersFromReader(header, reader, size)

	return storage.NewObjectPartialContent(header, reader), nil
}

func handleError(err error) (*storage.ObjectReader, error) {
	if !errors.Is(err, gcs.ErrBucketNotExist) && !errors.Is(err, gcs.ErrObjectNotExist) {
		return nil, err
	}

	return storage.NewObjectNotFound(err.Error()), nil
}

func setHeadersFromReader(header http.Header, reader *gcs.Reader, size int64) {
	header.Set(httpheaders.AcceptRanges, "bytes")
	header.Set(httpheaders.ContentLength, strconv.Itoa(int(size)))
	header.Set(httpheaders.ContentType, reader.Attrs.ContentType)
	header.Set(httpheaders.CacheControl, reader.Attrs.CacheControl)
}
