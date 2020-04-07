package main

import (
	"net/http"
	"regexp"
	"strings"

	nanoid "github.com/matoous/go-nanoid"
)

const (
	xRequestIDHeader = "X-Request-ID"
)

var (
	requestIDRe = regexp.MustCompile(`^[A-Za-z0-9_\-]+$`)
)

type routeHandler func(string, http.ResponseWriter, *http.Request)
type panicHandler func(string, http.ResponseWriter, *http.Request, error)

type route struct {
	Method  string
	Prefix  string
	Handler routeHandler
	Exact   bool
}

type router struct {
	prefix       string
	Routes       []*route
	PanicHandler panicHandler
}

func (r *route) IsMatch(req *http.Request) bool {
	if r.Method != req.Method {
		return false
	}

	if r.Exact {
		return req.URL.Path == r.Prefix
	}

	return strings.HasPrefix(req.URL.Path, r.Prefix)
}

func newRouter(prefix string) *router {
	return &router{
		prefix: prefix,
		Routes: make([]*route, 0),
	}
}

func (r *router) Add(method, prefix string, handler routeHandler, exact bool) {
	r.Routes = append(
		r.Routes,
		&route{Method: method, Prefix: r.prefix + prefix, Handler: handler, Exact: exact},
	)
}

func (r *router) GET(prefix string, handler routeHandler, exact bool) {
	r.Add(http.MethodGet, prefix, handler, exact)
}

func (r *router) OPTIONS(prefix string, handler routeHandler, exact bool) {
	r.Add(http.MethodOptions, prefix, handler, exact)
}

func (r *router) HEAD(prefix string, handler routeHandler, exact bool) {
	r.Add(http.MethodHead, prefix, handler, exact)
}

func (r *router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req = req.WithContext(setTimerSince(req.Context()))

	reqID := req.Header.Get(xRequestIDHeader)

	if len(reqID) == 0 || !requestIDRe.MatchString(reqID) {
		reqID, _ = nanoid.Nanoid()
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

	logRequest(reqID, req)

	for _, rr := range r.Routes {
		if rr.IsMatch(req) {
			rr.Handler(reqID, rw, req)
			return
		}
	}

	logWarning("Route for %s is not defined", req.URL.Path)

	rw.WriteHeader(404)
}
