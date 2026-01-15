package stream

import (
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/imgproxy/imgproxy/v3/cookies"
	"github.com/imgproxy/imgproxy/v3/errctx"
	"github.com/imgproxy/imgproxy/v3/fetcher"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/monitoring"
	"github.com/imgproxy/imgproxy/v3/options"
	"github.com/imgproxy/imgproxy/v3/options/keys"
	"github.com/imgproxy/imgproxy/v3/server"
)

const (
	streamBufferSize     = 4096        // Size of the buffer used for streaming
	errCategoryStreaming = "streaming" // Streaming error category
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

// HandlerContext provides access to shared handler dependencies
type HandlerContext interface {
	Fetcher() *fetcher.Fetcher
	Cookies() *cookies.Cookies
	Monitoring() *monitoring.Monitoring
}

// Handler handles image passthrough requests, allowing images to be streamed directly
type Handler struct {
	HandlerContext

	config *Config // Configuration for the streamer
}

// request holds the parameters and state for a single streaming request
type request struct {
	HandlerContext

	handler      *Handler
	imageRequest *http.Request
	imageURL     string
	reqID        string
	opts         *options.Options
	rw           server.ResponseWriter
}

// New creates new handler object
func New(hCtx HandlerContext, config *Config) (*Handler, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Handler{
		HandlerContext: hCtx,
		config:         config,
	}, nil
}

// Execute handles the image passthrough request, streaming the image directly to the response writer
func (s *Handler) Execute(
	userRequest *http.Request,
	imageURL string,
	reqID string,
	o *options.Options,
	rw server.ResponseWriter,
) *server.Error {
	stream := &request{
		HandlerContext: s.HandlerContext,
		handler:        s,
		imageRequest:   userRequest,
		imageURL:       imageURL,
		reqID:          reqID,
		opts:           o,
		rw:             rw,
	}

	return stream.execute()
}

// execute handles the actual streaming logic
func (s *request) execute() *server.Error {
	s.Monitoring().Stats().IncImagesInProgress()
	defer s.Monitoring().Stats().DecImagesInProgress()

	ctx, cancelSpan := s.Monitoring().StartSpan(s.imageRequest.Context(), "Streaming image", nil)
	defer cancelSpan()

	// Passthrough request headers from the original request
	requestHeaders := s.getImageRequestHeaders()
	cookieJar, err := s.getCookieJar()
	if err != nil {
		return s.wrapError(err)
	}

	// Build the request to fetch the image
	r, err := s.Fetcher().BuildRequest(ctx, s.imageURL, requestHeaders, cookieJar)
	if r != nil {
		defer r.Cancel()
	}
	if err != nil {
		return s.wrapError(err)
	}

	// Send the request to fetch the image
	res, err := r.Send()
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return s.wrapError(err)
	}

	// Output streaming response headers
	s.rw.SetOriginHeaders(res.Header)
	s.rw.Passthrough(s.handler.config.PassthroughResponseHeaders...) // NOTE: priority? This is lowest as it was
	s.rw.SetContentLength(int(res.ContentLength))
	s.rw.SetCanonical(s.imageURL)
	s.rw.SetExpires(s.opts.GetTime(keys.Expires))

	// Set the Content-Disposition header
	s.setContentDisposition(r.URL().Path, res)

	// Copy the status code from the original response
	s.rw.WriteHeader(res.StatusCode)

	// Write the actual data
	s.streamData(res)

	return nil
}

// getCookieJar returns non-empty cookie jar if cookie passthrough is enabled
func (s *request) getCookieJar() (http.CookieJar, error) {
	return s.Cookies().JarFromRequest(s.imageRequest)
}

// getImageRequestHeaders returns a new http.Header containing only
// the headers that should be passed through from the user request
func (s *request) getImageRequestHeaders() http.Header {
	h := make(http.Header)
	httpheaders.CopyFromRequest(s.imageRequest, h, s.handler.config.PassthroughRequestHeaders)

	return h
}

// setContentDisposition writes the headers to the response writer
func (s *request) setContentDisposition(imagePath string, serverResponse *http.Response) {
	// Try to set correct Content-Disposition file name and extension
	if serverResponse.StatusCode < 200 || serverResponse.StatusCode >= 300 {
		return
	}

	ct := serverResponse.Header.Get(httpheaders.ContentType)

	s.rw.SetContentDisposition(
		imagePath,
		s.opts.GetString(keys.Filename, ""),
		"",
		ct,
		s.opts.GetBool(keys.ReturnAttachment, false),
	)
}

// streamData copies the image data from the response body to the response writer
func (s *request) streamData(res *http.Response) {
	buf := streamBufPool.Get().(*[]byte) //nolint:forcetypeassert
	defer streamBufPool.Put(buf)

	_, copyerr := io.CopyBuffer(s.rw, res.Body, *buf)

	server.LogResponse(
		s.reqID, s.imageRequest, res.StatusCode, nil,
		slog.String("image_url", s.imageURL),
		slog.Any("processing_options", s.opts),
	)

	// We've got to skip logging here
	if copyerr != nil {
		panic(http.ErrAbortHandler)
	}
}

func (s *request) wrapError(err error) *server.Error {
	return server.NewError(errctx.WrapWithStackSkip(err, 1), errCategoryStreaming)
}
