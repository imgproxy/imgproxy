//go:build integration
// +build integration

// Integration test helpers for imgproxy.
// We use regular `go build` instead of Docker to make sure
// tests run in the same environment as other tests,
// including in CI, where everything runs in a custom Docker image
// against the different libvips versions.

package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	buildContext = ".."                 // Source code folder
	binPath      = "/tmp/imgproxy-test" // Path to the built imgproxy binary
	bindPort     = 9090                 // Port to bind imgproxy to
	bindHost     = "127.0.0.1"          // Host to bind imgproxy to
)

var (
	buildCmd = []string{"build", "-v", "-ldflags=-s -w", "-o", binPath} // imgproxy build command
)

// waitForPort tries to connect to host:port until successful or timeout
func waitForPort(host string, port int, timeout time.Duration) error {
	var address string
	if net.ParseIP(host) != nil && net.ParseIP(host).To4() == nil {
		// IPv6 address, wrap in brackets
		address = fmt.Sprintf("[%s]:%d", host, port)
	} else {
		address = fmt.Sprintf("%s:%d", host, port)
	}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
		if err == nil {
			conn.Close()
			return nil // port is open
		}
		time.Sleep(200 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for port %s", address)
}

func startImgproxy(t *testing.T, ctx context.Context, testImagesPath string) string {
	// Build the imgproxy binary
	buildCmd := exec.Command("go", buildCmd...)
	buildCmd.Dir = buildContext
	buildCmd.Env = os.Environ()
	buildOut, err := buildCmd.CombinedOutput()
	require.NoError(t, err, "failed to build imgproxy: %v\n%s", err, string(buildOut))

	// Start imgproxy in the background
	cmd := exec.CommandContext(ctx, binPath)

	// Set environment variables for imgproxy
	cmd.Env = append(os.Environ(), "IMGPROXY_BIND=:"+fmt.Sprintf("%d", bindPort))
	cmd.Env = append(cmd.Env, "IMGPROXY_LOCAL_FILESYSTEM_ROOT="+testImagesPath)
	cmd.Env = append(cmd.Env, "IMGPROXY_MAX_ANIMATION_FRAMES=999")
	cmd.Env = append(cmd.Env, "IMGPROXY_VIPS_LEAK_CHECK=true")
	cmd.Env = append(cmd.Env, "IMGPROXY_LOG_MEM_STATS=true")
	cmd.Env = append(cmd.Env, "IMGPROXY_DEVELOPMENT_ERRORS_MODE=true")

	// That one is for the build logs
	stdout, _ := os.CreateTemp("", "imgproxy-stdout-*")
	stderr, _ := os.CreateTemp("", "imgproxy-stderr-*")
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err = cmd.Start()
	require.NoError(t, err, "failed to start imgproxy: %v", err)

	// Wait for port 8090 to be available
	err = waitForPort(bindHost, bindPort, 5*time.Second)
	if err != nil {
		cmd.Process.Kill()
		require.NoError(t, err, "imgproxy did not start in time")
	}

	// Return a dummy container (nil) and connection string
	t.Cleanup(func() {
		cmd.Process.Kill()
		stdout.Close()
		stderr.Close()
		os.Remove(stdout.Name())
		os.Remove(stderr.Name())
		os.Remove(binPath)
	})

	return fmt.Sprintf("%s:%d", bindHost, bindPort)
}

// fetchImage fetches an image from the imgproxy server
func fetchImage(t *testing.T, cs string, path string) []byte {
	url := fmt.Sprintf("http://%s/%s", cs, path)

	resp, err := http.Get(url)
	require.NoError(t, err, "Failed to fetch image from %s", url)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200 OK, got %d, url: %s", resp.StatusCode, url)

	bytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Failed to read response body from %s", url)

	return bytes
}

// testImagesPath returns the absolute path to the test images directory
func testImagesPath(t *testing.T) (string, error) {
	// Get current working directory
	dir, err := os.Getwd()
	require.NoError(t, err)

	// Convert to absolute path (if it's not already)
	absPath, err := filepath.Abs(dir)
	require.NoError(t, err)

	return path.Join(absPath, "../testdata/test-images"), nil
}
