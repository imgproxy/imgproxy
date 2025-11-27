package integration

import (
	"os"
	"testing"

	"github.com/imgproxy/imgproxy/v3"
)

// TestMain performs global setup/teardown for the integration tests.
func TestMain(m *testing.M) {
	err := imgproxy.Init()
	if err != nil {
		panic(err)
	}

	r := m.Run()
	imgproxy.Shutdown()
	os.Exit(r)
}
