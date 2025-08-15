// router represents our HTTP server routes
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
	xAmznRequestContextHeader = "x-amzn-request-context"
)

var (
	requestIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
)

// RouteHandler represents error handler which might return an error
type RouteHandler func(string, http.ResponseWriter, *http.Request) error

// route represents route
type route struct {
	Method  string
	Prefix  string
	Handler RouteHandler
	Exact   bool
}

// Router represents current server router
type Router struct {
	config       *Config
	prefix       string
	healthRoutes []string
	faviconRoute string

	Routes        []*route
	HealthHandler RouteHandler
}

// NewRouter creates a new router
func NewRouter(config *Config) *Router {
	prefix := config.PathPrefix

	healthRoutes := []string{prefix + "/health"}
	if len(config.HealthCheckPath) > 0 {
		healthRoutes = append(healthRoutes, prefix+config.HealthCheckPath)
	}

	return &Router{
		config:       config,
		prefix:       prefix,
		healthRoutes: healthRoutes,
		faviconRoute: prefix + "/favicon.ico",
		Routes:       nil,
	}
}

// Add adds new route to the set
func (r *Router) Add(method, prefix string, handler RouteHandler, exact bool) {
	// Don't add routes with empty prefix
	if len(r.prefix+prefix) == 0 {
		return
	}

	r.Routes = append(
		r.Routes,
		&route{Method: method, Prefix: r.prefix + prefix, Handler: handler, Exact: exact},
	)
}

func (r *Router) GET(prefix string, handler RouteHandler, exact bool) {
	r.Add(http.MethodGet, prefix, handler, exact)
}

func (r *Router) OPTIONS(prefix string, handler RouteHandler, exact bool) {
	r.Add(http.MethodOptions, prefix, handler, exact)
}

func (r *Router) HEAD(prefix string, handler RouteHandler, exact bool) {
	r.Add(http.MethodHead, prefix, handler, exact)
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req, timeoutCancel := startRequestTimer(req)
	defer timeoutCancel()

	rw = newTimeoutResponse(rw)

	reqID := req.Header.Get(httpheaders.XRequestID)

	if len(reqID) == 0 || !requestIDRe.MatchString(reqID) {
		if lambdaContextVal := req.Header.Get(xAmznRequestContextHeader); len(lambdaContextVal) > 0 {
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

	rw.Header().Set(httpheaders.Server, "imgproxy")
	rw.Header().Set(httpheaders.XRequestID, reqID)

	if req.Method == http.MethodGet {
		if r.HealthHandler != nil {
			if slices.Contains(r.healthRoutes, req.URL.Path) {
				r.HealthHandler(reqID, rw, req)
				return
			}
		}

		if req.URL.Path == r.faviconRoute {
			// TODO: Add a real favicon maybe?
			rw.Header().Set(httpheaders.ContentType, "text/plain")
			rw.WriteHeader(404)
			// Write a single byte to make AWS Lambda happy
			rw.Write([]byte{' '})
			return
		}
	}

	if ip := req.Header.Get("CF-Connecting-IP"); len(ip) != 0 {
		replaceRemoteAddr(req, ip)
	} else if ip := req.Header.Get(httpheaders.XForwardedFor); len(ip) != 0 {
		if index := strings.Index(ip, ","); index > 0 {
			ip = ip[:index]
		}
		replaceRemoteAddr(req, ip)
	} else if ip := req.Header.Get(httpheaders.XRealIP); len(ip) != 0 {
		replaceRemoteAddr(req, ip)
	}

	LogRequest(reqID, req)

	for _, rr := range r.Routes {
		if rr.isMatch(req) {
			rr.Handler(reqID, rw, req)
			return
		}
	}

	LogResponse(reqID, req, 404, newRouteNotDefinedError(req.URL.Path))

	rw.Header().Set(httpheaders.ContentType, "text/plain")
	rw.WriteHeader(http.StatusNotFound)
	rw.Write([]byte{' '})
}

func replaceRemoteAddr(req *http.Request, ip string) {
	_, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		port = "80"
	}

	req.RemoteAddr = net.JoinHostPort(strings.TrimSpace(ip), port)
}

func (r *route) isMatch(req *http.Request) bool {
	if r.Method != req.Method {
		return false
	}

	if r.Exact {
		return req.URL.Path == r.Prefix
	}

	return strings.HasPrefix(req.URL.Path, r.Prefix)
}
