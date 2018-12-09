package bugsnag

import (
	"context"
	"net/http"
	"strings"
)

const requestContextKey requestKey = iota

type requestKey int

// AttachRequestData returns a child of the given context with the request
// object attached for later extraction by the notifier in order to
// automatically record request data
func AttachRequestData(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, requestContextKey, r)
}

// extractRequestInfo looks for the request object that the notifier
// automatically attaches to the context when using any of the supported
// frameworks or bugsnag.HandlerFunc or bugsnag.Handler, and returns sub-object
// supported by the notify API.
func extractRequestInfo(ctx context.Context) (*RequestJSON, *http.Request) {
	if req := getRequestIfPresent(ctx); req != nil {
		return extractRequestInfoFromReq(req), req
	}
	return nil, nil
}

// extractRequestInfoFromReq extracts the request information the notify API
// understands from the given HTTP request. Returns the sub-object supported by
// the notify API.
func extractRequestInfoFromReq(req *http.Request) *RequestJSON {
	proto := "http://"
	if req.TLS != nil {
		proto = "https://"
	}
	return &RequestJSON{
		ClientIP:   req.RemoteAddr,
		HTTPMethod: req.Method,
		URL:        proto + req.Host + req.RequestURI,
		Referer:    req.Referer(),
		Headers:    parseRequestHeaders(req.Header),
	}
}

func parseRequestHeaders(header map[string][]string) map[string]string {
	headers := make(map[string]string)
	for k, v := range header {
		// Headers can have multiple values, in which case we report them as csv
		if contains(Config.ParamsFilters, k) {
			headers[k] = "[FILTERED]"
		} else {
			headers[k] = strings.Join(v, ",")
		}
	}
	return headers
}

func contains(slice []string, e string) bool {
	for _, s := range slice {
		if strings.ToLower(s) == strings.ToLower(e) {
			return true
		}
	}
	return false
}

func getRequestIfPresent(ctx context.Context) *http.Request {
	if ctx == nil {
		return nil
	}
	val := ctx.Value(requestContextKey)
	if val == nil {
		return nil
	}
	return val.(*http.Request)
}
