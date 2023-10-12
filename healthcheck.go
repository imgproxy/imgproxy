package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"

	"github.com/imgproxy/imgproxy/v3/config"
	"github.com/imgproxy/imgproxy/v3/config/configurators"
)

func healthcheck() int {
	network := config.Network
	bind := config.Bind
	pathprefix := config.PathPrefix

	configurators.String(&network, "IMGPROXY_NETWORK")
	configurators.String(&bind, "IMGPROXY_BIND")
	configurators.URLPath(&pathprefix, "IMGPROXY_PATH_PREFIX")

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
		return 1
	}
	defer res.Body.Close()

	msg, _ := io.ReadAll(res.Body)
	fmt.Fprintln(os.Stderr, string(msg))

	if res.StatusCode != 200 {
		return 1
	}

	return 0
}
