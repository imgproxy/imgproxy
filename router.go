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
}

type router struct {
	Routes       []*route
	PanicHandler panicHandler
}

func newRouter() *router {
	return &router{
		Routes: make([]*route, 0),
	}
}

func (r *router) Add(method, prefix string, handler routeHandler) {
	r.Routes = append(
		r.Routes,
		&route{Method: method, Prefix: prefix, Handler: handler},
	)
}

func (r *router) GET(prefix string, handler routeHandler) {
	r.Add(http.MethodGet, prefix, handler)
}

func (r *router) OPTIONS(prefix string, handler routeHandler) {
	r.Add(http.MethodOptions, prefix, handler)
}

func (r *router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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
		if rr.Method == req.Method && strings.HasPrefix(req.URL.Path, rr.Prefix) {
			rr.Handler(reqID, rw, req)
			return
		}
	}

	logWarning("Route for %s is not defined", req.URL.Path)

	rw.WriteHeader(404)
}
