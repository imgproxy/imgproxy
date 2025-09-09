package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/imgproxy/imgproxy/v3"
	"github.com/imgproxy/imgproxy/v3/httpheaders"
	"github.com/imgproxy/imgproxy/v3/testutil"
)

type TestServer struct {
	Addr     net.Addr
	Shutdown context.CancelFunc
}

// Suite is a test suite for integration tests.
//
// It lazily initializes [imgproxy.Config] and [imgproxy.Imgproxy] when they are accessed.
//
// It provides the [Suite.GET] method that lazily initializes a test imgproxy server
// and performs a GET request against it.
//
// Take note that Suite utilizes SetupSuite and TearDownSuite for setup and cleanup.
// If you define them for your test suite, make sure to call the base methods.
type Suite struct {
	testutil.LazySuite

	TestData *testutil.TestDataProvider

	Config   testutil.LazyObj[*imgproxy.Config]
	Imgproxy testutil.LazyObj[*imgproxy.Imgproxy]
	Server   testutil.LazyObj[*TestServer]
}

func (s *Suite) SetupSuite() {
	// Silence all the logs
	logrus.SetOutput(io.Discard)

	// Initialize test data provider (local test files)
	s.TestData = testutil.NewTestDataProvider(s.T)

	s.Config, _ = testutil.NewLazySuiteObj(s, func() (*imgproxy.Config, error) {
		// TODO: replace with NewDefaultConfig when we get rid of global config
		c, err := imgproxy.LoadConfigFromEnv(nil)
		s.Require().NoError(err)

		c.Server.Bind = ":0"

		c.Fetcher.Transport.Local.Root = s.TestData.Root()
		c.Fetcher.Transport.HTTP.ClientKeepAliveTimeout = 0

		return c, nil
	})

	s.Imgproxy, _ = testutil.NewLazySuiteObj(s, func() (*imgproxy.Imgproxy, error) {
		return imgproxy.New(s.T().Context(), s.Config())
	})

	s.Server, _ = testutil.NewLazySuiteObj(
		s,
		func() (*TestServer, error) {
			return s.startServer(s.Imgproxy()), nil
		},
		func(s *TestServer) error {
			s.Shutdown()
			return nil
		},
	)
}

func (s *Suite) TearDownSuite() {
	logrus.SetOutput(os.Stdout)
}

// startServer starts imgproxy instance's server for the tests.
// Returns [TestServer] that contains the server address and shutdown function
func (s *Suite) startServer(i *imgproxy.Imgproxy) *TestServer {
	ctx, cancel := context.WithCancel(s.T().Context())

	addrCh := make(chan net.Addr)

	go func() {
		err := i.StartServer(ctx, addrCh)
		if err != nil {
			s.T().Errorf("Imgproxy stopped with error: %v", err)
		}
	}()

	return &TestServer{
		Addr:     <-addrCh,
		Shutdown: cancel,
	}
}

// GET performs a GET request to the imageproxy real server
func (s *Suite) GET(path string, header ...http.Header) *http.Response {
	url := fmt.Sprintf("http://%s%s", s.Server().Addr, path)

	// Perform GET request to an url
	req, err := http.NewRequest("GET", url, nil)
	s.Require().NoError(err)

	// Copy headers from the provided http.Header to the request
	for _, h := range header {
		httpheaders.CopyAll(h, req.Header, true)
	}

	// Do the request
	resp, err := http.DefaultClient.Do(req)
	s.Require().NoError(err)

	return resp
}
