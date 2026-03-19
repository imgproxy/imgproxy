package integration_test

import (
	"testing"

	"github.com/imgproxy/imgproxy/v3/testutil/servertest"
)

// TestMain performs global setup/teardown for the integration tests.
func TestMain(m *testing.M) {
	servertest.TestMain(m)
}
