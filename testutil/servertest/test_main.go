package servertest

import (
	"context"
	"os"
	"testing"

	"github.com/imgproxy/imgproxy/v3"
)

// TestMain performs global setup/teardown for the integration tests.
// Use it in packages that use [Suite] for integration tests.
func TestMain(m *testing.M) {
	err := imgproxy.Init(context.Background())
	if err != nil {
		panic(err)
	}

	r := m.Run()
	imgproxy.Shutdown()
	os.Exit(r)
}
