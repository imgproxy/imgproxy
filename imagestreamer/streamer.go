// imagestreamer is responsible for handling image passthrough streaming
package imagestreamer

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/handlererr"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/metrics"
	"github.com/imgproxy/imgproxy/v3/metrics/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/router"
	"github.com/imgproxy/imgproxy/v3/stemext"
	log "github.com/sirupsen/logrus"
)

const (
	streamBufferSize = 4096 // Size of the buffer used for streaming
)

var (
	// streamBufPool is a sync.Pool for reusing byte slices used for streaming
	streamBufPool = sync.Pool{
		New: func() any {
			buf := make([]byte, streamBufferSize)
			return &buf
		},
	}
)

// Request represents an image request that will be processed by the image streamer
// NOTE: This struct will be used as a base for the image request in the processing handler.
// Here it's temporary.
type Request struct {
	UserRequest       *http.Request              // Original user request to imgproxy
	ImageURL          string                     // URL of the image to be streamed
	ReqID             string                     // Unique identifier for the request
	ProcessingOptions *options.ProcessingOptions // Processing options for the image
}

// streamer handles image passthrough requests, allowing images to be streamed directly
type streamer struct {
	fetcher             *imagefetcher.Fetcher // Fetcher instance to handle image fetching
	headerWriterFactory *headerwriter.Factory // Factory for creating header writers
	config              *Config               // Configuration for the streamer
	p                   *Request              // Streaming request
	rw                  http.ResponseWriter   // Response writer to write the streamed image
}

// Stream handles the image passthrough request, streaming the image directly to the response writer
func (s *streamer) Stream(ctx context.Context) {
	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()
	defer metrics.StartStreamingSegment(ctx)()

	// Passthrough request headers from the original request
	requestHeaders := s.getPassthroughRequestHeaders()
	cookieJar := s.getCookieJar(ctx)

	// Build the request to fetch the image
	r, err := s.fetcher.BuildRequest(ctx, s.p.ImageURL, requestHeaders, cookieJar)
	defer r.Cancel()
	handlererr.Check(ctx, handlererr.ErrTypeStreaming, err)

	// Send the request to fetch the image
	res, err := r.Send()
	if res != nil {
		defer res.Body.Close()
	}
	handlererr.Check(ctx, handlererr.ErrTypeStreaming, err)

	s.writeHeaders(r, res)
	s.sendData(res)
}

// getCookieJar returns non-empty cookie jar if cookie passthrough is enabled
func (s *streamer) getCookieJar(ctx context.Context) http.CookieJar {
	if !s.config.CookiePassthrough {
		return nil
	}

	cookieJar, err := cookies.JarFromRequest(s.p.UserRequest)
	handlererr.Check(ctx, handlererr.ErrTypeStreaming, err)

	return cookieJar
}

// getPassthroughRequestHeaders returns a new http.Header containing only
// the headers that should be passed through from the user request
func (s *streamer) getPassthroughRequestHeaders() http.Header {
	h := make(http.Header)

	for _, key := range s.config.PassthroughRequestHeaders {
		values := s.p.UserRequest.Header.Values(key)

		for _, value := range values {
			h.Add(key, value)
		}
	}

	return h
}

// writeHeaders writes the headers to the response writer
func (s *streamer) writeHeaders(r *imagefetcher.Request, res *http.Response) {
	hw := s.headerWriterFactory.NewHeaderWriter(s.rw.Header(), r.URL().String())

	// Copy the response headers to the header writer
	hw.Copy(s.config.KeepResponseHeaders)

	// Set the Content-Length header
	hw.SetContentLength(int(res.ContentLength))

	// Try to set correct Content-Disposition file name and extension
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		ct := res.Header.Get("Content-Type")

		// Try to best guess the file name and extension
		stem, ext := stemext.FromURL(r.URL()).
			SetExtFromContentTypeIfEmpty(ct).
			OverrideStem(s.p.ProcessingOptions.Filename).
			StemExt()

		// Write the Content-Disposition header
		hw.SetContentDisposition(stem, ext, s.p.ProcessingOptions.ReturnAttachment)
	}

	hw.Write(s.rw)
}

// sendData copies the image data from the response body to the response writer
func (s *streamer) sendData(res *http.Response) {
	buf := streamBufPool.Get().(*[]byte)
	defer streamBufPool.Put(buf)

	_, copyerr := io.CopyBuffer(s.rw, res.Body, *buf)

	router.LogResponse(
		s.p.ReqID, s.p.UserRequest, res.StatusCode, nil,
		log.Fields{
			"image_url":          s.p.ImageURL,
			"processing_options": s.p.ProcessingOptions,
		},
	)

	if copyerr != nil {
		panic(http.ErrAbortHandler)
	}
}
