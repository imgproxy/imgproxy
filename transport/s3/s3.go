package s3

import (
	"fmt"
	http "net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/imgproxy/imgproxy/v3/config"
)

// transport implements RoundTripper for the 's3' protocol.
type transport struct {
	svc *s3.S3
}

func New() (http.RoundTripper, error) {
	s3Conf := aws.NewConfig()

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

	if config.ETagEnabled {
		if ifNoneMatch := req.Header.Get("If-None-Match"); len(ifNoneMatch) > 0 {
			input.IfNoneMatch = aws.String(ifNoneMatch)
		}
	}

	s3req, _ := t.svc.GetObjectRequest(input)

	if err := s3req.Send(); err != nil {
		return nil, err
	}

	return s3req.HTTPResponse, nil
}
