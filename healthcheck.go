package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"

	"github.com/imgproxy/imgproxy/v2/config"
	"github.com/imgproxy/imgproxy/v2/config/configurators"
)

func healthcheck() int {
	network := config.Network
	bind := config.Bind

	configurators.String(&network, "IMGPROXY_NETWORK")
	configurators.String(&bind, "IMGPROXY_BIND")

	httpc := http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial(network, bind)
			},
		},
	}

	res, err := httpc.Get("http://imgproxy/health")
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	defer res.Body.Close()

	msg, _ := ioutil.ReadAll(res.Body)
	fmt.Fprintln(os.Stderr, string(msg))

	if res.StatusCode != 200 {
		return 1
	}

	return 0
}
