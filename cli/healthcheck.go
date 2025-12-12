package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/urfave/cli/v3"
)

// healthcheck performs a healthcheck on a running imgproxy instance
func healthcheck(ctx context.Context, c *cli.Command) error {
	var network, bind, pathprefix string

	err := errors.Join(
		server.IMGPROXY_NETWORK.Parse(&network),
		server.IMGPROXY_BIND.Parse(&bind),
		server.IMGPROXY_PATH_PREFIX.Parse(&pathprefix),
	)
	if err != nil {
		return err
	}

	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial(network, bind)
			},
		},
	}

	res, err := httpc.Get(fmt.Sprintf("http://imgproxy%s/health", pathprefix))
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
