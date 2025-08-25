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

// Handler handles image passthrough requests, allowing images to be streamed directly
type Handler struct {
	config  *Config               // Configuration for the streamer
	fetcher *imagefetcher.Fetcher // Fetcher instance to handle image fetching
	hw      *headerwriter.Writer  // Configured HeaderWriter instance
}

// request holds the parameters and state for a single streaming request
type request struct {
	handler      *Handler
	imageRequest *http.Request
	imageURL     string
	reqID        string
	po           *options.ProcessingOptions
	rw           http.ResponseWriter
}

// New creates new handler object
func New(config *Config, hw *headerwriter.Writer, fetcher *imagefetcher.Fetcher) (*Handler, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Handler{
		fetcher: fetcher,
		config:  config,
		hw:      hw,
	}, nil
}

// Stream handles the image passthrough request, streaming the image directly to the response writer
func (s *Handler) Execute(
	ctx context.Context,
	userRequest *http.Request,
	imageURL string,
	reqID string,
	po *options.ProcessingOptions,
	rw http.ResponseWriter,
) error {
	stream := &request{
		handler:      s,
		imageRequest: userRequest,
		imageURL:     imageURL,
		reqID:        reqID,
		po:           po,
		rw:           rw,
	}

	return stream.execute(ctx)
}

// execute handles the actual streaming logic
func (s *request) execute(ctx context.Context) error {
	stats.IncImagesInProgress()
	defer stats.DecImagesInProgress()
	defer monitoring.StartStreamingSegment(ctx)()

	// Passthrough request headers from the original request
	requestHeaders := s.getImageRequestHeaders()
	cookieJar, err := s.getCookieJar()
	if err != nil {
		return ierrors.Wrap(err, 0, ierrors.WithCategory(categoryStreaming))
	}

	// Build the request to fetch the image
	r, err := s.handler.fetcher.BuildRequest(ctx, s.imageURL, requestHeaders, cookieJar)
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
	hw := s.handler.hw.NewRequest(res.Header, s.imageURL)

	hw.Passthrough(s.handler.config.PassthroughResponseHeaders) // NOTE: priority? This is lowest as it was
	hw.SetContentLength(int(res.ContentLength))
	hw.SetCanonical()
	hw.SetExpires(s.po.Expires)
	hw.Write(s.rw)

	// Write Content-Disposition header
	s.writeContentDisposition(r.URL().Path, res)

	// Copy the status code from the original response
	s.rw.WriteHeader(res.StatusCode)

	// Write the actual data
	s.streamData(res)

	return nil
}

// getCookieJar returns non-empty cookie jar if cookie passthrough is enabled
func (s *request) getCookieJar() (http.CookieJar, error) {
	if !s.handler.config.CookiePassthrough {
		return nil, nil
	}

	return cookies.JarFromRequest(s.imageRequest)
}

// getImageRequestHeaders returns a new http.Header containing only
// the headers that should be passed through from the user request
func (s *request) getImageRequestHeaders() http.Header {
	h := make(http.Header)

	for _, key := range s.handler.config.PassthroughRequestHeaders {
		values := s.imageRequest.Header.Values(key)

		for _, value := range values {
			h.Add(key, value)
		}
	}

	return h
}

// writeContentDisposition writes the headers to the response writer
func (s *request) writeContentDisposition(imagePath string, serverResponse *http.Response) {
	// Try to set correct Content-Disposition file name and extension
	if serverResponse.StatusCode < 200 || serverResponse.StatusCode >= 300 {
		return
	}

	ct := serverResponse.Header.Get(httpheaders.ContentType)

	// Try to best guess the file name and extension
	cd := httpheaders.ContentDispositionValue(
		imagePath,
		s.po.Filename,
		"",
		ct,
		s.po.ReturnAttachment,
	)

	// Write the Content-Disposition header
	s.rw.Header().Set(httpheaders.ContentDisposition, cd)
}

// streamData copies the image data from the response body to the response writer
func (s *request) streamData(res *http.Response) {
	buf := streamBufPool.Get().(*[]byte)
	defer streamBufPool.Put(buf)

	_, copyerr := io.CopyBuffer(s.rw, res.Body, *buf)

	server.LogResponse(
		s.reqID, s.imageRequest, res.StatusCode, nil,
		log.Fields{
			"image_url":          s.imageURL,
			"processing_options": s.po,
		},
	)

	// We've got to skip logging here
	if copyerr != nil {
		panic(http.ErrAbortHandler)
	}
}
