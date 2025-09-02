package integration

import (
	"os"
	"testing"

	"github.com/imgproxy/imgproxy/v3"
)

const (
	bindPort = 9090 // Port to bind imgproxy to
	bindHost = "localhost"
)

// TestMain performs global setup/teardown for the integration tests.
func TestMain(m *testing.M) {
	imgproxy.Init()
	os.Exit(m.Run())
	imgproxy.Shutdown()
}
