package main

import (
	"bytes"
	"fmt"
	"github.com/ncw/swift"
	"io/ioutil"
	"net/http"
	"strings"
)

type swiftTransport struct {
	swiftConnection swift.Connection
}

func newSwiftTransport() swiftTransport {
	var (
		swiftConnection swift.Connection
	)
	swiftConnection = swift.Connection{
		UserName: conf.SwiftUserName,
		ApiKey:   conf.SwiftPassword,
		AuthUrl:  conf.SwiftAuthUrl,
		Tenant:   conf.SwiftTenant,
	}

	err := swiftConnection.Authenticate()
	if err != nil {
		fmt.Println("Error authenticating with Swift")
		fmt.Println(err)
		panic(err)
	}

	return swiftTransport{swiftConnection: swiftConnection}
}

func (t swiftTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	imgPath := strings.Split(req.URL.Path, "/")

	body, err := t.swiftConnection.ObjectGetBytes(req.URL.Host, imgPath[1])

	if err != nil {
		fmt.Println("Error getting img from Swift")
		fmt.Println(err)
		return nil, err
	}

	return &http.Response{
		Status:        "200 OK",
		StatusCode:    200,
		Proto:         "HTTP/1.0",
		ProtoMajor:    1,
		ProtoMinor:    0,
		Header:        make(http.Header),
		ContentLength: int64(len(body)),
		Body:          ioutil.NopCloser(bytes.NewReader(body)),
		Close:         true,
		Request:       req,
	}, nil
}
