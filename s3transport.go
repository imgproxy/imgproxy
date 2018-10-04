package main

import (
	"fmt"
	http "net/http"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// s3Transport implements RoundTripper for the 's3' protocol.
type s3Transport struct {
	svc *s3.S3
}

func newS3Transport() http.RoundTripper {
	return s3Transport{s3.New(session.New())}
}

func (t s3Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(req.URL.Host),
		Key:    aws.String(req.URL.Path),
	}
	s3req, _ := t.svc.GetObjectRequest(input)

	s3err := s3req.Send()
	if s3err == nil { // resp is now filled
		return s3req.HTTPResponse, nil
	}
	fmt.Println("s3 error", s3err)
	return nil, s3err
}
