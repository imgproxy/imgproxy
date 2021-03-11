package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"net/http"
	"path"
	"strings"
	"os"
)

var pushdPath = "/pushd"
var s3ImagesBucket = os.Getenv("PUSH_S3_IMAGES_BUCKET")
var s3RenderBucket = os.Getenv("PUSH_S3_RENDER_BUCKET")

func fileNameToParams(requestUri string) string {
	imgParams := make(map[string]string)

	splitFilename := strings.Split(path.Base(requestUri), "__")
	for _, param := range splitFilename {
		splitParam := strings.Split(param, "_")
		if len(splitParam) < 2{
			continue
		}
		imgParams[splitParam[0]] = splitParam[1]
	}

	urlParam := fmt.Sprintf("/plain/s3://%s/%s@jpg", s3ImagesBucket, splitFilename[len(splitFilename)-1])

	var pathStr string
	for param, value := range imgParams {
		paramString := fmt.Sprintf("/%s:%s", param, value)
		pathStr = pathStr + paramString
	}
	// if no signature required, add a filler signature
	var paramPath string
	if conf.AllowInsecure{
		paramPath = "/nosig" + pathStr + urlParam
	} else {
		paramPath = pathStr + urlParam
	}
	logNotice("parsed pushd file name to: %s", paramPath)
	return paramPath

}

func getCachePath(requestUri string) string {
	pathBase := path.Base(requestUri)
	pathDirs := strings.Split(path.Dir(requestUri), "/")
	// only add last pathDir if length > 2, we expect at least /pushd in the path
	if len(pathDirs) > 2 {
		return pathBase
	} else {
		return fmt.Sprintf("%s/%s", pathDirs[len(pathDirs)], pathBase)
	}
}

func beforeProcessing(r *http.Request) (*http.Request, string){
	// process pushd filename if path starts with pushdPath
	var cachePath string
	if strings.HasPrefix(r.RequestURI, pushdPath) {
		cachePath = getCachePath(r.RequestURI)
		logNotice("Will upload with S3 key: %s")
		r.RequestURI = fileNameToParams(r.RequestURI)
	}
	return r, cachePath
}

func uploadToS3(data []byte, s3Key string) {
	svc := s3.New(session.New())
	input := &s3.PutObjectInput{
		Body:    aws.ReadSeekCloser(bytes.NewReader(data)),
		Bucket:  aws.String(s3RenderBucket),
		Key:     aws.String(s3Key),
	}

	_, err := svc.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				logError(aerr.Error())
			}
		} else {
			logError(err.Error())
		}
		return
	}
}

func beforeResponse(imageData []byte, cachePath string) {
	logNotice("I would be uploading to S3 here, if I was implemented yet")
	uploadToS3(imageData, cachePath)
}
