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

	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
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

	config *Config
}

func New(config *Config, trans *http.Transport) (http.RoundTripper, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	conf, err := awsConfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("can't load AWS S3 config"))
	}

	conf.HTTPClient = &http.Client{Transport: trans}

	if len(config.Region) != 0 {
		conf.Region = config.Region
	}

	if len(conf.Region) == 0 {
		conf.Region = "us-west-1"
	}

	if len(config.AssumeRoleArn) != 0 {
		creds := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(conf), config.AssumeRoleArn, func(o *stscreds.AssumeRoleOptions) {
			if len(config.AssumeRoleExternalID) != 0 {
				o.ExternalID = aws.String(config.AssumeRoleExternalID)
			}
		})
		conf.Credentials = creds
	}

	clientOptions := []func(*s3.Options){
		func(o *s3.Options) {
			o.DisableLogOutputChecksumValidationSkipped = true
		},
	}

	if len(config.Endpoint) != 0 {
		endpoint := config.Endpoint
		if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
			endpoint = "http://" + endpoint
		}
		clientOptions = append(clientOptions, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = config.EndpointUsePathStyle
		})
	}

	client, err := createClient(conf, clientOptions, config)
	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("can't create S3 client"))
	}

	return &transport{
		clientOptions:   clientOptions,
		defaultClient:   client,
		defaultConfig:   conf,
		clientsByRegion: map[string]s3Client{conf.Region: client},
		clientsByBucket: make(map[string]s3Client),
		config:          config,
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
			Header:        http.Header{httpheaders.ContentType: {"text/plain"}},
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
		if ifNoneMatch := req.Header.Get("If-None-Match"); len(ifNoneMatch) > 0 {
			input.IfNoneMatch = aws.String(ifNoneMatch)
		}

		if ifModifiedSince := req.Header.Get("If-Modified-Since"); len(ifModifiedSince) > 0 {
			parsedIfModifiedSince, err := time.Parse(http.TimeFormat, ifModifiedSince)
			if err == nil {
				input.IfModifiedSince = &parsedIfModifiedSince
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

	if t.config.DecryptionClientEnabled {
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

	client, err := createClient(conf, t.clientOptions, t.config)
	if err != nil {
		return nil, ierrors.Wrap(err, 0, ierrors.WithPrefix("can't create regional S3 client"))
	}

	t.clientsByRegion[region] = client
	t.clientsByBucket[bucket] = client

	return client, nil
}

func createClient(conf aws.Config, opts []func(*s3.Options), config *Config) (s3Client, error) {
	client := s3.NewFromConfig(conf, opts...)

	if config.DecryptionClientEnabled {
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
