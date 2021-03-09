package main

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"os"
)

var pushdPath = "/pushd"
var s3Bucket = os.Getenv("PUSH_S3_BUCKET")

func fileNameToParams(fileName string) string {
	imgParams := make(map[string]string)

	splitFilename := strings.Split(path.Base(fileName), "__")
	for _, param := range splitFilename {
		splitParam := strings.Split(param, "_")
		if len(splitParam) < 2{
			continue
		}
		imgParams[splitParam[0]] = splitParam[1]
	}

	urlParam := fmt.Sprintf("/plain/s3://%s/%s@jpg", s3Bucket, splitFilename[len(splitFilename)-1])

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

func beforeProcessing(r *http.Request) *http.Request{
	// process pushd filename if path starts with pushdPath
	if strings.HasPrefix(r.RequestURI, pushdPath) {
		r.RequestURI = fileNameToParams(r.RequestURI)
	}
	return r
}

func beforeResponse(imageData []byte) {
	logNotice("I would be uploading to S3 here, if I was implemented yet")
}
