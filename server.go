package main

import (
	"bytes"
	"compress/gzip"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type httpHandler struct {
	sem chan struct{}
}

func newHttpHandler() httpHandler {
	return httpHandler{make(chan struct{}, conf.Concurrency)}
}

func parsePath(r *http.Request) (string, processingOptions, error) {
	var po processingOptions
	var err error

	path := r.URL.Path
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")

	if len(parts) < 7 {
		return "", po, errors.New("Invalid path")
	}

	token := parts[0]

	if err = validatePath(token, strings.TrimPrefix(path, fmt.Sprintf("/%s", token))); err != nil {
		return "", po, err
	}

	po.resize = parts[1]

	if po.width, err = strconv.Atoi(parts[2]); err != nil {
		return "", po, fmt.Errorf("Invalid width: %s", parts[2])
	}

	if po.height, err = strconv.Atoi(parts[3]); err != nil {
		return "", po, fmt.Errorf("Invalid height: %s", parts[3])
	}

	if g, ok := gravityTypes[parts[4]]; ok {
		po.gravity = g
	} else {
		return "", po, fmt.Errorf("Invalid gravity: %s", parts[4])
	}

	po.enlarge = parts[5] != "0"

	filenameParts := strings.Split(strings.Join(parts[6:], ""), ".")

	if len(filenameParts) < 2 {
		po.format = imageTypes["jpg"]
	} else if f, ok := imageTypes[filenameParts[1]]; ok {
		po.format = f
	} else {
		return "", po, fmt.Errorf("Invalid image format: %s", filenameParts[1])
	}

	filename, err := base64.RawURLEncoding.DecodeString(filenameParts[0])
	if err != nil {
		return "", po, errors.New("Invalid filename encoding")
	}

	return string(filename), po, nil
}

func imageContentType(b []byte) string {
	_, imgtype, _ := image.DecodeConfig(bytes.NewReader(b))
	return fmt.Sprintf("image/%s", imgtype)
}

func logResponse(status int, msg string) {
	var color int

	if status > 500 {
		color = 31
	} else if status > 400 {
		color = 33
	} else {
		color = 32
	}

	log.Printf("|\033[7;%dm %d \033[0m| %s\n", color, status, msg)
}

func respondWithImage(r *http.Request, rw http.ResponseWriter, data []byte, imgURL string, po processingOptions, startTime time.Time) {
	gzipped := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && conf.GZipCompression > 0

	rw.Header().Set("Expires", time.Now().Add(time.Second*time.Duration(conf.TTL)).Format(http.TimeFormat))
	rw.Header().Set("Cache-Control", fmt.Sprintf("Cache-Control: max-age=%d", conf.TTL))
	rw.Header().Set("Content-Type", imageContentType(data))
	if gzipped {
		rw.Header().Set("Content-Encoding", "gzip")
	}

	rw.WriteHeader(200)

	if gzipped {
		gz, _ := gzip.NewWriterLevel(rw, conf.GZipCompression)
		gz.Write(data)
		gz.Close()
	} else {
		rw.Write(data)
	}

	logResponse(200, fmt.Sprintf("Processed in %s: %s; %+v", time.Since(startTime), imgURL, po))
}

func respondWithError(rw http.ResponseWriter, status int, err error, msg string) {
	logResponse(status, err.Error())

	rw.WriteHeader(status)
	rw.Write([]byte(msg))
}

func repondWithForbidden(rw http.ResponseWriter) {
	logResponse(403, "Invalid secret")

	rw.WriteHeader(403)
	rw.Write([]byte("Forbidden"))
}

func checkSecret(s string) bool {
	if len(conf.Secret) == 0 {
		return true
	}
	return strings.HasPrefix(s, "Bearer ") && subtle.ConstantTimeCompare([]byte(strings.TrimPrefix(s, "Bearer ")), []byte(conf.Secret)) == 1
}

func (h *httpHandler) lock() {
	h.sem <- struct{}{}
}

func (h *httpHandler) unlock() {
	defer func() { <-h.sem }()
}

func (h httpHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	log.Printf("GET: %s\n", r.URL.RequestURI())

	h.lock()
	defer h.unlock()

	t := time.Now()

	if !checkSecret(r.Header.Get("Authorization")) {
		repondWithForbidden(rw)
		return
	}

	imgURL, procOpt, err := parsePath(r)
	if err != nil {
		respondWithError(rw, 404, err, "Invalid image url")
		return
	}

	if _, err = url.ParseRequestURI(imgURL); err != nil {
		respondWithError(rw, 404, err, "Invalid image url")
		return
	}

	b, err := downloadImage(imgURL)
	if err != nil {
		respondWithError(rw, 404, err, "Image is unreacable")
		return
	}

	b, err = processImage(b, procOpt)
	if err != nil {
		respondWithError(rw, 500, err, "Error occured while processing image")
		return
	}

	respondWithImage(r, rw, b, imgURL, procOpt, t)
}
