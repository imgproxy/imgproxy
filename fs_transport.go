package main

import (
	"fmt"
	"net/http"
)

type fsTransport struct {
	fs http.Dir
}

func newFsTransport() fsTransport {
	return fsTransport{fs: http.Dir(conf.LocalFileSystemRoot)}
}

func (t fsTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	f, err := t.fs.Open(req.URL.Path)

	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("%s is a directory", req.URL.Path)
	}

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        make(http.Header),
		ContentLength: fi.Size(),
		Body:          f,
		Close:         true,
		Request:       req,
	}, nil
}
