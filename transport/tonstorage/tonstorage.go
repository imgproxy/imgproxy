package tonstorage

import (
	"context"
	"errors"
	"github.com/imgproxy/imgproxy/v3/config"
	"net/http"
	"time"
)

// transport implements RoundTripper for the 's3' protocol.
type transport struct {
	gateway string
	client  *http.Client
}

func New(client *http.Client) http.RoundTripper {
	return transport{config.TonstorageGateway, client}
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	for i := 0; i < 5; i++ {
		newReq, err := http.NewRequest("GET", t.gateway+req.URL.Host+req.URL.RequestURI(), nil)
		if err != nil {
			return nil, err
		}
		resp, err = t.client.Do(newReq.WithContext(req.Context()))
		if err == nil {
			return resp, nil
		}
		if errors.Is(err, context.Canceled) {
			return nil, err
		}
		time.Sleep(time.Second)
	}
	return resp, err
}
