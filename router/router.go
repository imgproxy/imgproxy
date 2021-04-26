package router

import (
	"net/http"
	"regexp"
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"
	log "github.com/sirupsen/logrus"
)

const (
	xRequestIDHeader = "X-Request-ID"
)

var (
	requestIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
)

type RouteHandler func(string, http.ResponseWriter, *http.Request)
type PanicHandler func(string, http.ResponseWriter, *http.Request, error)

type route struct {
	Method  string
	Prefix  string
	Handler RouteHandler
	Exact   bool
}

type Router struct {
	prefix       string
	Routes       []*route
	PanicHandler PanicHandler
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
	return &Router{
		prefix: prefix,
		Routes: make([]*route, 0),
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
	req = setRequestTime(req)

	reqID := req.Header.Get(xRequestIDHeader)

	if len(reqID) == 0 || !requestIDRe.MatchString(reqID) {
		reqID, _ = nanoid.New()
	}

	rw.Header().Set("Server", "imgproxy")
	rw.Header().Set(xRequestIDHeader, reqID)

	defer func() {
		if rerr := recover(); rerr != nil {
			if err, ok := rerr.(error); ok && r.PanicHandler != nil {
				r.PanicHandler(reqID, rw, req, err)
			} else {
				panic(rerr)
			}
		}
	}()

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
