package s3

import (
	"fmt"
	"io"
	http "net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/imgproxy/imgproxy/v3/config"
	defaultTransport "github.com/imgproxy/imgproxy/v3/transport"
)

// transport implements RoundTripper for the 's3' protocol.
type transport struct {
	svc *s3.S3
}

func New() (http.RoundTripper, error) {
	s3Conf := aws.NewConfig()

	trans, err := defaultTransport.New(false)
	if err != nil {
		return nil, err
	}

	s3Conf.HTTPClient = &http.Client{Transport: trans}

	if len(config.S3Region) != 0 {
		s3Conf.Region = aws.String(config.S3Region)
	}

	if len(config.S3Endpoint) != 0 {
		s3Conf.Endpoint = aws.String(config.S3Endpoint)
		s3Conf.S3ForcePathStyle = aws.Bool(true)
	}

	sess, err := session.NewSession()
	if err != nil {
		return nil, fmt.Errorf("Can't create S3 session: %s", err)
	}

	if len(config.S3AssumeRoleArn) != 0 {
		s3Conf.Credentials = stscreds.NewCredentials(sess, config.S3AssumeRoleArn)
	}

	if sess.Config.Region == nil || len(*sess.Config.Region) == 0 {
		sess.Config.Region = aws.String("us-west-1")
	}

	return transport{s3.New(sess, s3Conf)}, nil
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
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

	s3req, _ := t.svc.GetObjectRequest(input)
	s3req.SetContext(req.Context())

	if err := s3req.Send(); err != nil {
		if s3req.HTTPResponse != nil && s3req.HTTPResponse.Body != nil {
			s3req.HTTPResponse.Body.Close()
		}

		if s3err, ok := err.(awserr.Error); ok && s3err.Code() == request.CanceledErrorCode {
			if e := s3err.OrigErr(); e != nil {
				return nil, e
			}
		}

		if s3err, ok := err.(awserr.RequestFailure); !ok || s3err.StatusCode() < 100 || s3err.StatusCode() == 301 {
			return nil, err
		} else {
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
				Request:       s3req.HTTPRequest,
			}, nil
		}
	}

	return s3req.HTTPResponse, nil
}
