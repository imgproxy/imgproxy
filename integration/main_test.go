package integration

import (
	"os"
	"testing"

	"github.com/imgproxy/imgproxy/v3"
)

// TestMain performs global setup/teardown for the integration tests.
func TestMain(m *testing.M) {
	imgproxy.Init()
	os.Exit(m.Run())
	imgproxy.Shutdown()
}
