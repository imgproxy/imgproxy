package s3

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// s3Client is an interface for S3 normal and crypto client
type s3Client interface {
	GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}
