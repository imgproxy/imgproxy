package s3

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/storage"
	"github.com/imgproxy/imgproxy/v3/storage/common"
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
			"invalid S3 Storage URL: bucket name or object key are empty",
		), nil
	}

	// Check if access to the container is allowed
	if !common.IsBucketAllowed(bucket, s.config.AllowedBuckets, s.config.DeniedBuckets) {
		return nil, fmt.Errorf("access to the S3 bucket %s is denied", bucket)
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if len(query) > 0 {
		input.VersionId = aws.String(query)
	}

	if r := reqHeader.Get(httpheaders.Range); len(r) != 0 {
		input.Range = aws.String(r)
	} else {
		if ifNoneMatch := reqHeader.Get(httpheaders.IfNoneMatch); len(ifNoneMatch) > 0 {
			input.IfNoneMatch = aws.String(ifNoneMatch)
		}

		if ifModifiedSince := reqHeader.Get(httpheaders.IfModifiedSince); len(ifModifiedSince) > 0 {
			parsedIfModifiedSince, err := time.Parse(http.TimeFormat, ifModifiedSince)
			if err == nil {
				input.IfModifiedSince = &parsedIfModifiedSince
			}
		}
	}

	output, _, err := callWithClient(s, bucket, func(client s3Client) (*s3.GetObjectOutput, error) {
		output, err := client.GetObject(ctx, input)

		defer func() {
			if err != nil && output != nil && output.Body != nil {
				output.Body.Close()
			}
		}()

		return output, err
	})

	if err != nil {
		return handleError(err)
	}

	contentLength := int64(-1)
	if output.ContentLength != nil {
		contentLength = *output.ContentLength
	}

	if s.config.DecryptionClientEnabled {
		if unencryptedContentLength := output.Metadata[httpheaders.XAmzMetaECL]; len(unencryptedContentLength) != 0 {
			cl, err := strconv.ParseInt(unencryptedContentLength, 10, 64)
			if err != nil {
				return handleError(err)
			}
			contentLength = cl
		}
	}

	header := make(http.Header)
	if contentLength > 0 {
		header.Set(httpheaders.ContentLength, strconv.FormatInt(contentLength, 10))
	}
	if output.ContentType != nil {
		header.Set(httpheaders.ContentType, *output.ContentType)
	}
	if output.ContentEncoding != nil {
		header.Set(httpheaders.ContentEncoding, *output.ContentEncoding)
	}
	if output.CacheControl != nil {
		header.Set(httpheaders.CacheControl, *output.CacheControl)
	}
	if output.ExpiresString != nil {
		header.Set(httpheaders.Expires, *output.ExpiresString)
	}
	if output.ETag != nil {
		header.Set(httpheaders.Etag, *output.ETag)
	}
	if output.LastModified != nil {
		header.Set(httpheaders.LastModified, output.LastModified.Format(http.TimeFormat))
	}
	if output.AcceptRanges != nil {
		header.Set(httpheaders.AcceptRanges, *output.AcceptRanges)
	}
	if output.ContentRange != nil {
		header.Set(httpheaders.ContentRange, *output.ContentRange)
		return storage.NewObjectPartialContent(header, output.Body), nil
	}

	return storage.NewObjectOK(header, output.Body), nil
}
