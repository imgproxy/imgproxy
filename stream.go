package main

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/server"
)

var (
	streamReqHeaders = []string{
		"If-None-Match",
		"If-Modified-Since",
		"Accept-Encoding",
		"Range",
	}

	streamRespHeaders = []string{
		"ETag",
		"Content-Type",
		"Content-Encoding",
		"Content-Range",
		"Accept-Ranges",
		"Last-Modified",
	}

	streamBufPool = sync.Pool{
		New: func() interface{} {
			buf := make([]byte, 4096)
			return &buf
		},
	}
)

func streamOriginImage(ctx context.Context, reqID string, r *http.Request, rw http.ResponseWriter, po *options.ProcessingOptions, imageURL string) {
	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()

	defer metrics.StartStreamingSegment(ctx)()

	var (
		cookieJar http.CookieJar
		err       error
	)

	imgRequestHeader := make(http.Header)

	for _, k := range streamReqHeaders {
		if v := r.Header.Get(k); len(v) != 0 {
			imgRequestHeader.Set(k, v)
		}
	}

	if config.CookiePassthrough {
		cookieJar, err = cookies.JarFromRequest(r)
		checkErr(ctx, "streaming", err)
	}

	req, err := imagedata.Fetcher.BuildRequest(r.Context(), imageURL, imgRequestHeader, cookieJar)
	defer req.Cancel()
	checkErr(ctx, "streaming", err)

	res, err := req.Send()
	if res != nil {
		defer res.Body.Close()
	}
	checkErr(ctx, "streaming", err)

	for _, k := range streamRespHeaders {
		vv := res.Header.Values(k)
		for _, v := range vv {
			rw.Header().Set(k, v)
		}
	}

	if res.ContentLength >= 0 {
		rw.Header().Set("Content-Length", strconv.Itoa(int(res.ContentLength)))
	}

	if res.StatusCode < 300 {
		contentDisposition := httpheaders.ContentDispositionValue(
			req.URL().Path,
			po.Filename,
			"",
			rw.Header().Get(httpheaders.ContentType),
			po.ReturnAttachment,
		)
		rw.Header().Set("Content-Disposition", contentDisposition)
	}

	setCacheControl(rw, po.Expires, res.Header)
	setCanonical(rw, imageURL)
	rw.Header().Set("Content-Security-Policy", "script-src 'none'")

	rw.WriteHeader(res.StatusCode)

	buf := streamBufPool.Get().(*[]byte)
	defer streamBufPool.Put(buf)

	_, copyerr := io.CopyBuffer(rw, res.Body, *buf)
	if copyerr == http.ErrBodyNotAllowed {
		// We can hit this for some statuses like 304 Not Modified.
		// We can ignore this error.
		copyerr = nil
	}

	server.LogResponse(
		reqID, r, res.StatusCode, nil,
		log.Fields{
			"image_url":          imageURL,
			"processing_options": po,
		},
	)

	if copyerr != nil {
		panic(http.ErrAbortHandler)
	}
}
