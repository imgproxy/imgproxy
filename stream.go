package main

import (
	"context"
	"io"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/imagedata"
	"github.com/imgproxy/imgproxy/v3/imagetype"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/router"
)

var (
	streamReqHeaders = []string{
		"If-None-Match",
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
		cookieJar *cookiejar.Jar
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

	req, reqCancel, err := imagedata.BuildImageRequest(r.Context(), imageURL, imgRequestHeader, cookieJar)
	defer reqCancel()
	checkErr(ctx, "streaming", err)

	res, err := imagedata.SendRequest(req)
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
		var filename, ext, mimetype string

		_, filename = filepath.Split(req.URL.Path)
		ext = filepath.Ext(filename)

		if len(po.Filename) > 0 {
			filename = po.Filename
		} else {
			filename = filename[:len(filename)-len(ext)]
		}

		mimetype = rw.Header().Get("Content-Type")

		if len(ext) == 0 && len(mimetype) > 0 {
			if exts, err := mime.ExtensionsByType(mimetype); err == nil && len(exts) != 0 {
				ext = exts[0]
			}
		}

		rw.Header().Set("Content-Disposition", imagetype.ContentDisposition(filename, ext, po.ReturnAttachment))
	}

	setCacheControl(rw, po.Expires, map[string]string{
		"Cache-Control": res.Header.Get("Cache-Control"),
		"Expires":       res.Header.Get("Expires"),
	})
	setCanonical(rw, imageURL)
	rw.Header().Set("Content-Security-Policy", "script-src 'none'")

	rw.WriteHeader(res.StatusCode)

	buf := streamBufPool.Get().(*[]byte)
	defer streamBufPool.Put(buf)

	_, copyerr := io.CopyBuffer(rw, res.Body, *buf)

	router.LogResponse(
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
