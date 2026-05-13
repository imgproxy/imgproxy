package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/imgproxy/imgproxy/v4/server"
	"github.com/urfave/cli/v3"
)

// healthcheck performs a healthcheck on a running imgproxy instance
func healthcheck(ctx context.Context, c *cli.Command) error {
	config, err := server.LoadConfigFromEnv(nil)
	if err != nil {
		return err
	}

	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial(config.Network, config.Bind)
			},
		},
	}

	res, err := httpc.Get(fmt.Sprintf("http://imgproxy%s/health", config.PathPrefix))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return cli.Exit(err, 1)
	}
	defer res.Body.Close()

	msg, _ := io.ReadAll(res.Body)
	fmt.Fprintln(os.Stderr, string(msg))

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("healthcheck failed: %s", msg)
		return cli.Exit(err, 1)
	}

	return nil
}
