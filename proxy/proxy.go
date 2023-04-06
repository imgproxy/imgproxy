package proxy

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
)

type Auth struct {
	Username string
	Password string
	Host     string
}

func Init(username, password, host string) proxy.Dialer {
	if username == "" || password == "" || host == "" {
		return nil
	}

	dialer, err := proxy.SOCKS5("tcp", host, &proxy.Auth{User: username, Password: password}, proxy.Direct)
	if err != nil {
		log.Fatalf("failed to init proxy server")
		return nil
	}

	return dialer
}
