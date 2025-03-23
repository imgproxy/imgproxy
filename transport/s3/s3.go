package s3

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	s3Crypto "github.com/aws/amazon-s3-encryption-client-go/v3/client"
	s3CryptoMaterials "github.com/aws/amazon-s3-encryption-client-go/v3/materials"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsHttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	defaultTransport "github.com/imgproxy/imgproxy/v3/transport"
	"github.com/imgproxy/imgproxy/v3/transport/common"
)

type s3Client interface {
	GetObject(ctx context.Context, input *s3.GetObjectInput, opts ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}

// transport implements RoundTripper for the 's3' protocol.
type transport struct {
	clientOptions []func(*s3.Options)

	defaultClient s3Client
	defaultConfig aws.Config

	clientsByRegion map[string]s3Client
	clientsByBucket map[string]s3Client

	mu sync.RWMutex
}

func New() (http.RoundTripper, error) {
	conf, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("can't load AWS S3 config"))
	}

	trans, err := defaultTransport.New(false)
	if err != nil {
		return nil, err
	}

	conf.HTTPClient = &http.Client{Transport: trans}

	if len(config.S3Region) != 0 {
		conf.Region = config.S3Region
	}

	if len(conf.Region) == 0 {
		conf.Region = "us-west-1"
	}

	if len(config.S3AssumeRoleArn) != 0 {
		creds := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(conf), config.S3AssumeRoleArn, func(o *stscreds.AssumeRoleOptions) {
			if len(config.S3AssumeRoleExternalID) != 0 {
				o.ExternalID = aws.String(config.S3AssumeRoleExternalID)
			}
		})
		conf.Credentials = creds
	}

	clientOptions := []func(*s3.Options){}

	if len(config.S3Endpoint) != 0 {
		endpoint := config.S3Endpoint
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			endpoint = "http://" + endpoint
		}
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = config.S3EndpointUsePathStyle
		})
	}

	client, err := createClient(conf, clientOptions)
	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("can't create S3 client"))
	}

	return &transport{
		clientOptions:   clientOptions,
		defaultClient:   client,
		defaultConfig:   conf,
		clientsByRegion: map[string]s3Client{conf.Region: client},
		clientsByBucket: make(map[string]s3Client),
	}, nil
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	bucket, key, query := common.GetBucketAndKey(req.URL)

	if len(bucket) == 0 || len(key) == 0 {
		body := strings.NewReader("Invalid S3 URL: bucket name or object key is empty")
		return &http.Response{
			StatusCode:    http.StatusNotFound,
			Proto:         "HTTP/1.0",
			ProtoMajor:    1,
			ProtoMinor:    0,
			Header:        http.Header{"Content-Type": {"text/plain"}},
			ContentLength: int64(body.Len()),
			Body:          io.NopCloser(body),
			Close:         false,
			Request:       req,
		}, nil
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if len(query) > 0 {
		input.VersionId = aws.String(query)
	}

	statusCode := http.StatusOK

	if r := req.Header.Get("Range"); len(r) != 0 {
		input.Range = aws.String(r)
	} else {
		if config.ETagEnabled {
			if ifNoneMatch := req.Header.Get("If-None-Match"); len(ifNoneMatch) > 0 {
				input.IfNoneMatch = aws.String(ifNoneMatch)
			}
		}
		if config.LastModifiedEnabled {
			if ifModifiedSince := req.Header.Get("If-Modified-Since"); len(ifModifiedSince) > 0 {
				parsedIfModifiedSince, err := time.Parse(http.TimeFormat, ifModifiedSince)
				if err == nil {
					input.IfModifiedSince = &parsedIfModifiedSince
				}
			}
		}
	}

	client := t.getBucketClient(bucket)

	output, err := client.GetObject(req.Context(), input)

	defer func() {
		if err != nil && output != nil && output.Body != nil {
			output.Body.Close()
		}
	}()

	if err != nil {
		// Check if the error is the region mismatch error.
		// If so, create a new client with the correct region and retry the request.
		if region := regionFromError(err); len(region) != 0 {
			client, err = t.createBucketClient(bucket, region)
			if err != nil {
				return handleError(req, err)
			}

			output, err = client.GetObject(req.Context(), input)
		}
	}

	if err != nil {
		return handleError(req, err)
	}

	contentLength := int64(-1)
	if output.ContentLength != nil {
		contentLength = *output.ContentLength
	}

	if config.S3DecryptionClientEnabled {
		if unencryptedContentLength := output.Metadata["X-Amz-Meta-X-Amz-Unencrypted-Content-Length"]; len(unencryptedContentLength) != 0 {
			cl, err := strconv.ParseInt(unencryptedContentLength, 10, 64)
			if err != nil {
				handleError(req, err)
			}
			contentLength = cl
		}
	}

	header := make(http.Header)
	if contentLength > 0 {
		header.Set("Content-Length", strconv.FormatInt(contentLength, 10))
	}
	if output.ContentType != nil {
		header.Set("Content-Type", *output.ContentType)
	}
	if output.ContentEncoding != nil {
		header.Set("Content-Encoding", *output.ContentEncoding)
	}
	if output.CacheControl != nil {
		header.Set("Cache-Control", *output.CacheControl)
	}
	if output.ExpiresString != nil {
		header.Set("Expires", *output.ExpiresString)
	}
	if output.ETag != nil {
		header.Set("ETag", *output.ETag)
	}
	if output.LastModified != nil {
		header.Set("Last-Modified", output.LastModified.Format(http.TimeFormat))
	}
	if output.AcceptRanges != nil {
		header.Set("Accept-Ranges", *output.AcceptRanges)
	}
	if output.ContentRange != nil {
		header.Set("Content-Range", *output.ContentRange)
		statusCode = http.StatusPartialContent
	}

	return &http.Response{
		StatusCode:    statusCode,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        header,
		ContentLength: contentLength,
		Body:          output.Body,
		Close:         true,
		Request:       req,
	}, nil
}

func (t *transport) getBucketClient(bucket string) s3Client {
	var client s3Client

	func() {
		t.mu.RLock()
		defer t.mu.RUnlock()
		client = t.clientsByBucket[bucket]
	}()

	if client != nil {
		return client
	}

	return t.defaultClient
}

func (t *transport) createBucketClient(bucket, region string) (s3Client, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Check again if someone did this before us
	if client := t.clientsByBucket[bucket]; client != nil {
		return client, nil
	}

	if client := t.clientsByRegion[region]; client != nil {
		t.clientsByBucket[bucket] = client
		return client, nil
	}

	conf := t.defaultConfig.Copy()
	conf.Region = region

	client, err := createClient(conf, t.clientOptions)
	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("can't create regional S3 client"))
	}

	t.clientsByRegion[region] = client
	t.clientsByBucket[bucket] = client

	return client, nil
}

func createClient(conf aws.Config, opts []func(*s3.Options)) (s3Client, error) {
	client := s3.NewFromConfig(conf, opts...)

	if config.S3DecryptionClientEnabled {
		kmsClient := kms.NewFromConfig(conf)
		keyring := s3CryptoMaterials.NewKmsDecryptOnlyAnyKeyKeyring(kmsClient)

		cmm, err := s3CryptoMaterials.NewCryptographicMaterialsManager(keyring)
		if err != nil {
			return nil, err
		}

		return s3Crypto.New(client, cmm)
	} else {
		return client, nil
	}
}

func regionFromError(err error) string {
	var rerr *awsHttp.ResponseError
	if !errors.As(err, &rerr) {
		return ""
	}

	if rerr.Response == nil || rerr.Response.StatusCode != 301 {
		return ""
	}

	return rerr.Response.Header.Get("X-Amz-Bucket-Region")
}

func handleError(req *http.Request, err error) (*http.Response, error) {
	var rerr *awsHttp.ResponseError
	if !errors.As(err, &rerr) {
		return nil, ierrors.Wrap(err, 0)
	}

	if rerr.Response == nil || rerr.Response.StatusCode < 100 || rerr.Response.StatusCode == 301 {
		return nil, ierrors.Wrap(err, 0)
	}

	return &http.Response{
		StatusCode:    rerr.Response.StatusCode,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        http.Header{"Content-Type": {"text/plain"}},
		ContentLength: int64(len(err.Error())),
		Body:          io.NopCloser(strings.NewReader(err.Error())),
		Close:         false,
		Request:       req,
	}, nil
}
