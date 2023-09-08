package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

var pushdPath = "/pushd"
var s3ImagesBucket = os.Getenv("PUSH_S3_IMAGES_BUCKET")
var s3RenderBucket = os.Getenv("PUSH_S3_RENDER_BUCKET")

func fileNameToParams(requestUri string, needsSig bool) string {
	imgParams := make(map[string]string)

	splitFilename := strings.Split(path.Base(requestUri), "__")
	for _, param := range splitFilename {
		splitParam := strings.Split(param, "_")
		if len(splitParam) < 2 {
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
	if needsSig {
		paramPath = pathStr + urlParam
	} else {
		paramPath = "/nosig" + pathStr + urlParam
	}
	log.Info("parsed pushd file name to: %s", paramPath)
	return paramPath

}

// Gets s3 path of source file
// removes pushdPath and uuid from path if they exist
func getS3SourcePath(requestUri string) string {
	pathDirs := strings.Split(path.Dir(requestUri), "/")
	s3PathDirs := []string{s3ImagesBucket}
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

func beforeProcessing(r *http.Request, needsSig bool) (*http.Request, string) {
	// process pushd filename if path starts with pushdPath
	var cachePath string
	if strings.HasPrefix(r.RequestURI, pushdPath) {
		cachePath = getS3CachePath(r.RequestURI)
		r.RequestURI = fileNameToParams(r.RequestURI, needsSig)
	}
	return r, cachePath
}

func createMD5Hash(data []byte) string {
	hash := md5.New()
	reader := bytes.NewReader(data)
	if _, err := io.Copy(hash, reader); err != nil {
		log.Error(err.Error())
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func uploadToS3(data []byte, s3Key string, uploaded chan bool) {
	md5Hash := createMD5Hash(data)
	log.Info("Uploading rendered image to: %s with md5 hash: %s", s3Key, md5Hash)
	awsSession, err := session.NewSession()
	if err != nil {
		log.Error(err.Error())
		return
	}
	svc := s3.New(awsSession)
	input := &s3.PutObjectInput{
		Body:        aws.ReadSeekCloser(bytes.NewReader(data)),
		Bucket:      aws.String(s3RenderBucket),
		Key:         aws.String(s3Key),
		ContentType: aws.String("image/jpeg"),
		ContentMD5:  aws.String(md5Hash),
	}

	_, err = svc.PutObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Error(aerr.Error())
			}
		} else {
			log.Error(err.Error())
		}
		return
	}
	log.Info("Upload complete for: %s", s3Key)
	uploaded <- true
}

func beforeResponse(imageData []byte, cachePath string) chan bool {
	uploaded := make(chan bool)
	go uploadToS3(imageData, cachePath, uploaded)
	return uploaded
}

func isValidUUID(u string) bool {
	_, err := uuid.Parse(u)
	return err == nil
}
