package httprange

import (
	"errors"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
)

func Parse(s string) (int64, int64, error) {
	if s == "" {
		return 0, 0, nil // header not present
	}

	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return 0, 0, errors.New("invalid range")
	}

	for _, ra := range strings.Split(s[len(b):], ",") {
		ra = textproto.TrimString(ra)
		if ra == "" {
			continue
		}

		i := strings.Index(ra, "-")
		if i < 0 {
			return 0, 0, errors.New("invalid range")
		}

		start, end := textproto.TrimString(ra[:i]), textproto.TrimString(ra[i+1:])

		if start == "" {
			// Don't support ranges without start since it looks like FFmpeg doen't use ones
			return 0, 0, errors.New("invalid range")
		}

		istart, err := strconv.ParseInt(start, 10, 64)
		if err != nil || i < 0 {
			return 0, 0, errors.New("invalid range")
		}

		var iend int64

		if end == "" {
			iend = -1
		} else {
			iend, err = strconv.ParseInt(end, 10, 64)
			if err != nil || istart > iend {
				return 0, 0, errors.New("invalid range")
			}
		}

		return istart, iend, nil
	}

	return 0, 0, errors.New("invalid range")
}

func InvalidHTTPRangeResponse(req *http.Request) *http.Response {
	return &http.Response{
		StatusCode:    http.StatusRequestedRangeNotSatisfiable,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        make(http.Header),
		ContentLength: 0,
		Body:          nil,
		Close:         false,
		Request:       req,
	}
}
