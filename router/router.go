package router

import (
	"net"
	"net/http"
	"regexp"
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"
	log "github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3/config"
)

const (
	xRequestIDHeader = "X-Request-ID"
)

var (
	requestIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
)

type RouteHandler func(string, http.ResponseWriter, *http.Request)

type route struct {
	Method  string
	Prefix  string
	Handler RouteHandler
	Exact   bool
}

type Router struct {
	prefix       string
	healthRoutes []string
	faviconRoute string

	Routes        []*route
	HealthHandler RouteHandler
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

func New(prefix string) *Router {
	healthRoutes := []string{prefix + "/health"}
	if len(config.HealthCheckPath) > 0 {
		healthRoutes = append(healthRoutes, prefix+config.HealthCheckPath)
	}

	return &Router{
		prefix:       prefix,
		healthRoutes: healthRoutes,
		faviconRoute: prefix + "/favicon.ico",
		Routes:       make([]*route, 0),
	}
}

func (r *Router) Add(method, prefix string, handler RouteHandler, exact bool) {
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

	reqID := req.Header.Get(xRequestIDHeader)

	if len(reqID) == 0 || !requestIDRe.MatchString(reqID) {
		reqID, _ = nanoid.New()
	}

	rw.Header().Set("Server", "imgproxy")
	rw.Header().Set(xRequestIDHeader, reqID)

	if req.Method == http.MethodGet {
		if r.HealthHandler != nil {
			for _, healthRoute := range r.healthRoutes {
				if req.URL.Path == healthRoute {
					r.HealthHandler(reqID, rw, req)
					return
				}
			}
		}

		if req.URL.Path == r.faviconRoute {
			// TODO: Add a real favicon maybe?
			rw.Header().Set("Content-Type", "text/plain")
			rw.WriteHeader(404)
			// Write a single byte to make AWS Lambda happy
			rw.Write([]byte{' '})
			return
		}
	}

	if ip := req.Header.Get("CF-Connecting-IP"); len(ip) != 0 {
		replaceRemoteAddr(req, ip)
	} else if ip := req.Header.Get("X-Forwarded-For"); len(ip) != 0 {
		if index := strings.Index(ip, ","); index > 0 {
			ip = ip[:index]
		}
		replaceRemoteAddr(req, ip)
	} else if ip := req.Header.Get("X-Real-IP"); len(ip) != 0 {
		replaceRemoteAddr(req, ip)
	}

	LogRequest(reqID, req)

	for _, rr := range r.Routes {
		if rr.isMatch(req) {
			rr.Handler(reqID, rw, req)
			return
		}
	}

	log.Warningf("Route for %s is not defined", req.URL.Path)

	rw.WriteHeader(404)
}

func replaceRemoteAddr(req *http.Request, ip string) {
	_, port, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		port = "80"
	}

	req.RemoteAddr = net.JoinHostPort(strings.TrimSpace(ip), port)
}
