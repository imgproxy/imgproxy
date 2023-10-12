package s3

import (
	"context"
	"fmt"
	"io"
	http "net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/imgproxy/imgproxy/v3/config"
	defaultTransport "github.com/imgproxy/imgproxy/v3/transport"
)

// transport implements RoundTripper for the 's3' protocol.
type transport struct {
	session       *session.Session
	defaultClient *s3.S3

	clientsByRegion map[string]*s3.S3
	clientsByBucket map[string]*s3.S3

	mu sync.RWMutex
}

func New() (http.RoundTripper, error) {
	conf := aws.NewConfig()

	trans, err := defaultTransport.New(false)
	if err != nil {
		return nil, err
	}

	conf.HTTPClient = &http.Client{Transport: trans}

	if len(config.S3Endpoint) != 0 {
		conf.Endpoint = aws.String(config.S3Endpoint)
		conf.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Can't create S3 session: %s", err)
	}

	if len(config.S3Region) != 0 {
		sess.Config.Region = aws.String(config.S3Region)
	}

	if sess.Config.Region == nil || len(*sess.Config.Region) == 0 {
		sess.Config.Region = aws.String("us-west-1")
	}

	if len(config.S3AssumeRoleArn) != 0 {
		conf.Credentials = stscreds.NewCredentials(sess, config.S3AssumeRoleArn)
	}

	client := s3.New(sess, conf)

	clientRegion := "us-west-1"
	if client.Config.Region != nil {
		clientRegion = *client.Config.Region
	}

	return &transport{
		session:       sess,
		defaultClient: client,

		clientsByRegion: map[string]*s3.S3{clientRegion: client},
		clientsByBucket: make(map[string]*s3.S3),
	}, nil
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(req.URL.Host),
		Key:    aws.String(req.URL.Path),
	}

	if len(req.URL.RawQuery) > 0 {
		input.VersionId = aws.String(req.URL.RawQuery)
	}

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

	client, err := t.getClient(req.Context(), *input.Bucket)
	if err != nil {
		return handleError(req, err)
	}

	s3req, _ := client.GetObjectRequest(input)
	s3req.SetContext(req.Context())

	if err := s3req.Send(); err != nil {
		if s3req.HTTPResponse != nil && s3req.HTTPResponse.Body != nil {
			s3req.HTTPResponse.Body.Close()
		}

		return handleError(req, err)
	}

	return s3req.HTTPResponse, nil
}

func (t *transport) getClient(ctx context.Context, bucket string) (*s3.S3, error) {
	if !config.S3MultiRegion {
		return t.defaultClient, nil
	}

	var client *s3.S3

	func() {
		t.mu.RLock()
		defer t.mu.RUnlock()
		client = t.clientsByBucket[bucket]
	}()

	if client != nil {
		return client, nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	// Check again if someone did this before us
	if client = t.clientsByBucket[bucket]; client != nil {
		return client, nil
	}

	region, err := s3manager.GetBucketRegionWithClient(ctx, t.defaultClient, bucket)
	if err != nil {
		return nil, err
	}

	if client = t.clientsByRegion[region]; client != nil {
		t.clientsByBucket[bucket] = client
		return client, nil
	}

	conf := t.defaultClient.Config.Copy()
	conf.Region = aws.String(region)

	client = s3.New(t.session, conf)

	t.clientsByRegion[region] = client
	t.clientsByBucket[bucket] = client

	return client, nil
}

func handleError(req *http.Request, err error) (*http.Response, error) {
	if s3err, ok := err.(awserr.Error); ok && s3err.Code() == request.CanceledErrorCode {
		if e := s3err.OrigErr(); e != nil {
			return nil, e
		}
	}

	s3err, ok := err.(awserr.RequestFailure)
	if !ok || s3err.StatusCode() < 100 || s3err.StatusCode() == 301 {
		return nil, err
	}

	body := strings.NewReader(s3err.Message())

	return &http.Response{
		StatusCode:    s3err.StatusCode(),
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        http.Header{},
		ContentLength: int64(body.Len()),
		Body:          io.NopCloser(body),
		Close:         false,
		Request:       req,
	}, nil
}
