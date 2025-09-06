package server

import (
	"encoding/json"
	"net"
	"net/http"
	"regexp"
	"slices"
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

const (
	// defaultServerName is the default name of the server
	defaultServerName = "imgproxy"
)

var (
	// requestIDRe is a regular expression for validating request IDs
	requestIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
)

// RouteHandler is a function that handles HTTP requests.
type RouteHandler func(string, http.ResponseWriter, *http.Request) error

// Middleware is a function that wraps a RouteHandler with additional functionality.
type Middleware func(next RouteHandler) RouteHandler

// route represents a single route in the router.
type route struct {
	method  string       // method is the HTTP method for a route
	path    string       // path represents a route path
	exact   bool         // exact means that path must match exactly, otherwise any prefixed matches
	handler RouteHandler // handler is the function that handles the route
	silent  bool         // Silent route (no logs)
}

// Router is responsible for routing HTTP requests
type Router struct {
	// config represents the server configuration
	config *Config

	// routes is the collection of all routes
	routes []*route
}

// NewRouter creates a new Router instance
func NewRouter(config *Config) (*Router, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &Router{config: config}, nil
}

// add adds an abitary route to the router
func (r *Router) add(method, path string, handler RouteHandler, middlewares ...Middleware) *route {
	for _, m := range middlewares {
		handler = m(handler)
	}

	exact := true
	if strings.HasSuffix(path, "*") {
		exact = false
		path = strings.TrimSuffix(path, "*")
	}

	newRoute := &route{
		method:  method,
		path:    r.config.PathPrefix + path,
		handler: handler,
		exact:   exact,
	}

	r.routes = append(r.routes, newRoute)

	// Sort routes by exact flag, exact routes go first in the
	// same order they were added
	slices.SortStableFunc(r.routes, func(a, b *route) int {
		switch {
		case a.exact == b.exact:
			return 0
		case a.exact:
			return -1
		default:
			return 1
		}
	})

	return newRoute
}

// GET adds GET route
func (r *Router) GET(path string, handler RouteHandler, middlewares ...Middleware) *route {
	return r.add(http.MethodGet, path, handler, middlewares...)
}

// OPTIONS adds OPTIONS route
func (r *Router) OPTIONS(path string, handler RouteHandler, middlewares ...Middleware) *route {
	return r.add(http.MethodOptions, path, handler, middlewares...)
}

// HEAD adds HEAD route
func (r *Router) HEAD(path string, handler RouteHandler, middlewares ...Middleware) *route {
	return r.add(http.MethodHead, path, handler, middlewares...)
}

// ServeHTTP serves routes
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Attach timer to the context
	req, timeoutCancel := startRequestTimer(req)
	defer timeoutCancel()

	// Create the response writer which times out on write
	rw = newTimeoutResponse(rw, r.config.WriteResponseTimeout)

	// Get/create request ID
	reqID := r.getRequestID(req)

	// Replace request IP from headers
	r.replaceRemoteAddr(req)

	rw.Header().Set(httpheaders.Server, defaultServerName)
	rw.Header().Set(httpheaders.XRequestID, reqID)

	for _, rr := range r.routes {
		if rr.isMatch(req) {
			// Set req.Pattern. We use it to trim path prefixes in handlers.
			req.Pattern = rr.path

			if !rr.silent {
				LogRequest(reqID, req)
			}

			rr.handler(reqID, rw, req)
			return
		}
	}

	// Means that we have not found matching route
	LogRequest(reqID, req)
	LogResponse(reqID, req, http.StatusNotFound, newRouteNotDefinedError(req.URL.Path))
	r.NotFoundHandler(reqID, rw, req)
}

// NotFoundHandler is default 404 handler
func (r *Router) NotFoundHandler(reqID string, rw http.ResponseWriter, req *http.Request) error {
	rw.Header().Set(httpheaders.ContentType, "text/plain")
	rw.WriteHeader(http.StatusNotFound)
	rw.Write([]byte{' '}) // Write a single byte to make AWS Lambda happy

	return nil
}

// OkHandler is a default 200 OK handler
func (r *Router) OkHandler(reqID string, rw http.ResponseWriter, req *http.Request) error {
	rw.Header().Set(httpheaders.ContentType, "text/plain")
	rw.WriteHeader(http.StatusOK)
	rw.Write([]byte{' '}) // Write a single byte to make AWS Lambda happy

	return nil
}

// getRequestID tries to read request id from headers or from lambda
// context or generates a new one if nothing found.
func (r *Router) getRequestID(req *http.Request) string {
	// Get request ID from headers (if any)
	reqID := req.Header.Get(httpheaders.XRequestID)

	if len(reqID) == 0 || !requestIDRe.MatchString(reqID) {
		lambdaContextVal := req.Header.Get(httpheaders.XAmznRequestContextHeader)

		if len(lambdaContextVal) > 0 {
			var lambdaContext struct {
				RequestID string `json:"requestId"`
			}

			err := json.Unmarshal([]byte(lambdaContextVal), &lambdaContext)
			if err == nil && len(lambdaContext.RequestID) > 0 {
				reqID = lambdaContext.RequestID
			}
		}
	}

	if len(reqID) == 0 || !requestIDRe.MatchString(reqID) {
		reqID, _ = nanoid.New()
	}

	return reqID
}

// replaceRemoteAddr rewrites the req.RemoteAddr property from request headers
func (r *Router) replaceRemoteAddr(req *http.Request) {
	cfConnectingIP := req.Header.Get(httpheaders.CFConnectingIP)
	xForwardedFor := req.Header.Get(httpheaders.XForwardedFor)
	xRealIP := req.Header.Get(httpheaders.XRealIP)

	switch {
	case len(cfConnectingIP) > 0:
		replaceRemoteAddr(req, cfConnectingIP)
	case len(xForwardedFor) > 0:
		if index := strings.Index(xForwardedFor, ","); index > 0 {
			xForwardedFor = xForwardedFor[:index]
		}
		replaceRemoteAddr(req, xForwardedFor)
	case len(xRealIP) > 0:
		replaceRemoteAddr(req, xRealIP)
	}
}

// replaceRemoteAddr sets the req.RemoteAddr for request
func replaceRemoteAddr(req *http.Request, ip string) {
	_, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		port = "80"
	}

	req.RemoteAddr = net.JoinHostPort(strings.TrimSpace(ip), port)
}

// isMatch checks that a request matches route
func (r *route) isMatch(req *http.Request) bool {
	methodMatches := r.method == req.Method
	notExactPathMathes := !r.exact && strings.HasPrefix(req.URL.Path, r.path)
	exactPathMatches := r.exact && req.URL.Path == r.path

	return methodMatches && (notExactPathMathes || exactPathMatches)
}

// Silent sets Silent flag which supresses logs to true. We do not need to log
// requests like /health of /favicon.ico
func (r *route) Silent() *route {
	r.silent = true
	return r
}
