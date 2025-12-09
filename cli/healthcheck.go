package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/imgproxy/imgproxy/v3/env"
	"github.com/imgproxy/imgproxy/v3/server"
	"github.com/urfave/cli/v3"
)

// healthcheck performs a healthcheck on a running imgproxy instance
func healthcheck(ctx context.Context, c *cli.Command) error {
	var network, bind, pathprefix string

	env.String(&network, server.IMGPROXY_NETWORK)
	env.String(&bind, server.IMGPROXY_BIND)
	env.String(&pathprefix, server.IMGPROXY_PATH_PREFIX)

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
