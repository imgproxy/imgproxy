package honeybadger

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Errors returned by the backend when unable to successfully handle payload.
var (
	ErrRateExceeded    = errors.New("Rate exceeded: slow down!")
	ErrPaymentRequired = errors.New("Payment required: expired trial or credit card?")
	ErrUnauthorized    = errors.New("Unauthorized: bad API key?")
)

func newServerBackend(config *Configuration) *server {
	return &server{
		URL:    &config.Endpoint,
		APIKey: &config.APIKey,
		Client: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   config.Timeout,
		},
		Timeout: &config.Timeout,
	}
}

type server struct {
	APIKey  *string
	URL     *string
	Timeout *time.Duration
	Client  *http.Client
}

func (s *server) Notify(feature Feature, payload Payload) error {
	// Copy the value from the pointer in case it has changed in the
	// configuration.
	s.Client.Timeout = *s.Timeout

	url, err := url.Parse(*s.URL)
	if err != nil {
		return err
	}
	url.Path = "v1/" + feature.Endpoint
	req, err := http.NewRequest("POST", url.String(), bytes.NewReader(payload.toJSON()))
	if err != nil {
		return err
	}

	req.Header.Set("X-API-Key", *s.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}()

	switch resp.StatusCode {
	case 201:
		return nil
	case 429, 503:
		return ErrRateExceeded
	case 402:
		return ErrPaymentRequired
	case 403:
		return ErrUnauthorized
	default:
		return fmt.Errorf(
			"request failed status=%d expected=%d",
			resp.StatusCode,
			http.StatusCreated)
	}
}
