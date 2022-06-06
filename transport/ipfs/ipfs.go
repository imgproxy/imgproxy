package ipfs

import (
	"github.com/imgproxy/imgproxy/v3/config"
	"net/http"
	"net/url"
)

// transport implements RoundTripper for the 's3' protocol.
type transport struct {
	gateway *url.URL
	client  *http.Client
}

func New(client *http.Client) (http.RoundTripper, error) {
	gateway, err := url.Parse(config.IPFSGateway)
	if err != nil {
		return nil, err
	}
	return transport{gateway, client}, nil
}

func (t transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	newReq, err := http.NewRequest("GET", t.gateway.Scheme+"://"+t.gateway.Host+"/"+req.URL.Scheme+"/"+req.URL.Host+req.URL.RequestURI(), nil)
	if err != nil {
		return nil, err
	}
	resp, err = t.client.Do(newReq.WithContext(req.Context()))
	return resp, err
}
