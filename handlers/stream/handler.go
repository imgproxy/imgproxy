package stream

import (
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/headerwriter"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/ierrors"
	"github.com/imgproxy/imgproxy/v3/imagefetcher"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/monitoring/stats"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/server"
	log "github.com/sirupsen/logrus"
)

const (
	streamBufferSize  = 4096        // Size of the buffer used for streaming
	categoryStreaming = "streaming" // Streaming error category
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

// StreamingParams represents an image request params that will be processed by the image streamer
type StreamingParams struct {
	UserRequest       *http.Request              // Original user request to imgproxy
	ImageURL          string                     // URL of the image to be streamed
	ReqID             string                     // Unique identifier for the request
	ProcessingOptions *options.ProcessingOptions // Processing options for the image
}

// Handler handles image passthrough requests, allowing images to be streamed directly
type Handler struct {
	fetcher  *imagefetcher.Fetcher // Fetcher instance to handle image fetching
	config   *Config               // Configuration for the streamer
	hwConfig *headerwriter.Config  // Configuration for header writing
	params   *StreamingParams      // Streaming request
	res      http.ResponseWriter   // Response writer to write the streamed image
}

// Stream handles the image passthrough request, streaming the image directly to the response writer
func (s *Handler) Execute(ctx context.Context) error {
	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()
	defer monitoring.StartStreamingSegment(ctx)()

	// Passthrough request headers from the original request
	requestHeaders := s.getPassthroughRequestHeaders()
	cookieJar, err := s.getCookieJar()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryStreaming))
	}

	// Build the request to fetch the image
	r, err := s.fetcher.BuildRequest(ctx, s.params.ImageURL, requestHeaders, cookieJar)
	defer r.Cancel()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryStreaming))
	}

	// Send the request to fetch the image
	res, err := r.Send()
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryStreaming))
	}

	// Output streaming response headers
	hw := headerwriter.New(s.hwConfig, res.Header, s.params.ImageURL)
	hw.Passthrough(s.config.KeepResponseHeaders) // NOTE: priority? This is lowest as it was
	hw.SetContentLength(int(res.ContentLength))
	hw.SetCanonical()
	hw.SetMaxAge(s.params.ProcessingOptions.Expires, 0)
	hw.Write(s.res)

	// Write Content-Disposition header
	s.writeContentDisposition(s.params.ImageURL, res)

	// Copy the status code from the original response
	s.res.WriteHeader(res.StatusCode)

	// Write the actual data
	s.streamData(res)

	return nil
}

// getCookieJar returns non-empty cookie jar if cookie passthrough is enabled
func (s *Handler) getCookieJar() (http.CookieJar, error) {
	if !s.config.CookiePassthrough {
		return nil, nil
	}

	return cookies.JarFromRequest(s.params.UserRequest)
}

// getPassthroughRequestHeaders returns a new http.Header containing only
// the headers that should be passed through from the user request
func (s *Handler) getPassthroughRequestHeaders() http.Header {
	h := make(http.Header)

	for _, key := range s.config.PassthroughRequestHeaders {
		values := s.params.UserRequest.Header.Values(key)

		for _, value := range values {
			h.Add(key, value)
		}
	}

	return h
}

// writeContentDisposition writes the headers to the response writer
func (s *Handler) writeContentDisposition(imagePath string, serverResponse *http.Response) {
	// Try to set correct Content-Disposition file name and extension
	if serverResponse.StatusCode >= 200 && serverResponse.StatusCode < 300 {
		ct := serverResponse.Header.Get(httpheaders.ContentType)
		po := s.params.ProcessingOptions

		// Try to best guess the file name and extension
		cd := httpheaders.ContentDispositionValue(
			imagePath,
			po.Filename,
			"",
			ct,
			po.ReturnAttachment,
		)

		// Write the Content-Disposition header
		s.res.Header().Set(httpheaders.ContentDisposition, cd)
	}
}

// streamData copies the image data from the response body to the response writer
func (s *Handler) streamData(res *http.Response) {
	buf := streamBufPool.Get().(*[]byte)
	defer streamBufPool.Put(buf)

	_, copyerr := io.CopyBuffer(s.res, res.Body, *buf)

	server.LogResponse(
		s.params.ReqID, s.params.UserRequest, res.StatusCode, nil,
		log.Fields{
			"image_url":          s.params.ImageURL,
			"processing_options": s.params.ProcessingOptions,
		},
	)

	// We've got to skip logging here
	if copyerr != nil {
		panic(http.ErrAbortHandler)
	}
}
