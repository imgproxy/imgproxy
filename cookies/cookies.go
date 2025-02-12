package cookies

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sync"

	"golang.org/x/net/publicsuffix"

	"github.com/imgproxy/imgproxy/v3/config"
)

type anyCookieJarEntry struct {
	Name   string
	Value  string
	Quoted bool
}

// anyCookieJar is a cookie jar that stores all cookies in memory
// and doesn't care about domains and paths
type anyCookieJar struct {
	entries []anyCookieJarEntry
	mu      sync.RWMutex
}

func (j *anyCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.mu.Lock()
	defer j.mu.Unlock()

	for _, c := range cookies {
		entry := anyCookieJarEntry{
			Name:   c.Name,
			Value:  c.Value,
			Quoted: c.Quoted,
		}
		j.entries = append(j.entries, entry)
	}
}

func (j *anyCookieJar) Cookies(u *url.URL) []*http.Cookie {
	j.mu.RLock()
	defer j.mu.RUnlock()

	cookies := make([]*http.Cookie, 0, len(j.entries))
	for _, e := range j.entries {
		c := http.Cookie{
			Name:   e.Name,
			Value:  e.Value,
			Quoted: e.Quoted,
		}
		cookies = append(cookies, &c)
	}

	return cookies
}

func JarFromRequest(r *http.Request) (jar http.CookieJar, err error) {
	if config.CookiePassthroughAll {
		jar = &anyCookieJar{}
	} else {
		jar, err = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			return nil, err
		}
	}

	if r == nil {
		return jar, nil
	}

	var cookieBase *url.URL

	if !config.CookiePassthroughAll {
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
				host = r.Host
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
	}

	jar.SetCookies(cookieBase, r.Cookies())

	return jar, nil
}
