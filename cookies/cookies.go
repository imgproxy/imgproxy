package cookies

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/publicsuffix"

	"github.com/imgproxy/imgproxy/v3/config"
)

func JarFromRequest(r *http.Request) (*cookiejar.Jar, error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	if r == nil {
		return jar, nil
	}

	var cookieBase *url.URL

	if len(config.CookieBaseURL) > 0 {
		if cookieBase, err = url.Parse(config.CookieBaseURL); err != nil {
			return nil, fmt.Errorf("can't parse cookie base URL: %s", err)
		}
	}

	if cookieBase == nil {
		scheme := r.Header.Get("X-Forwarded-Proto")
		if len(scheme) == 0 {
			scheme = "http"
		}

		host := r.Header.Get("X-Forwarded-Host")
		if len(host) == 0 {
			host = r.Header.Get("Host")
		}

		if len(host) == 0 {
			return jar, nil
		}

		port := r.Header.Get("X-Forwarded-Port")
		if len(port) > 0 {
			host = host + ":" + port
		}

		cookieBase = &url.URL{
			Scheme: scheme,
			Host:   host,
		}
	}

	jar.SetCookies(cookieBase, r.Cookies())

	return jar, nil
}
