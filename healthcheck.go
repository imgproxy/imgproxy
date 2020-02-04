package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

func healthcheck() {
	network := conf.Network
	bind := conf.Bind

	strEnvConfig(&network, "IMGPROXY_NETWORK")
	strEnvConfig(&bind, "IMGPROXY_BIND")

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
		os.Exit(1)
	}
	defer res.Body.Close()

	msg, _ := ioutil.ReadAll(res.Body)
	fmt.Fprintln(os.Stderr, string(msg))

	if res.StatusCode != 200 {
		os.Exit(1)
	}

	os.Exit(0)
}
