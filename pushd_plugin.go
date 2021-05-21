package main

import (
	"bytes"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"net/http"
	"os"
	"path"
	"strings"
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

	s3Path := getS3SourcePath(requestUri)

	urlParam := fmt.Sprintf("/plain/s3://%s/%s@jpg", s3Path, splitFilename[len(splitFilename)-1])

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

// Gets s3 path of source file
// removes pushdPath and uuid from path if they exist
func getS3SourcePath(requestUri string) string {
	pathDirs := strings.Split(path.Dir(requestUri), "/")
	s3PathDirs := []string{ s3ImagesBucket }
	for _, pathDir := range pathDirs {
		if len(pathDir) < 1 || isValidUUID(pathDir) || pathDir == pushdPath[1:] {
			continue
		} else {
			s3PathDirs = append(s3PathDirs, pathDir)
		}
	}
	return strings.Join(s3PathDirs, "/")
}

// creates s3 path for cached generated file
func getS3CachePath(requestUri string) string {
	pathBase := path.Base(requestUri)
	pathDirs := strings.Split(path.Dir(requestUri), "/")
	// only add last pathDir if length > 2, we expect at least /pushd in the path
	if len(pathDirs) <= 2 {
		return pathBase
	} else {
		return fmt.Sprintf("%s/%s", strings.Join(pathDirs[2:], "/"), pathBase)
	}
}

func beforeProcessing(r *http.Request) (*http.Request, string){
	// process pushd filename if path starts with pushdPath
	var cachePath string
	if strings.HasPrefix(r.RequestURI, pushdPath) {
		cachePath = getS3CachePath(r.RequestURI)
		r.RequestURI = fileNameToParams(r.RequestURI)
	}
	return r, cachePath
}

func uploadToS3(data []byte, s3Key string, uploaded chan bool) {
	svc := s3.New(session.New())
	input := &s3.PutObjectInput{
		Body:    aws.ReadSeekCloser(bytes.NewReader(data)),
		Bucket:  aws.String(s3RenderBucket),
		Key:     aws.String(s3Key),
		ContentType: aws.String("image/jpeg"),
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
	logNotice("Upload complete for: %s", s3Key)
	uploaded <- true
}

func beforeResponse(imageData []byte, cachePath string) chan bool {
	logNotice("Uploading rendered image to: %s", cachePath)
	uploaded := make(chan bool)
	go uploadToS3(imageData, cachePath, uploaded)
	return uploaded
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
