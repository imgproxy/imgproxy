package cookies

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"slices"
	"sync"

	"golang.org/x/net/publicsuffix"

	"github.com/imgproxy/imgproxy/v3/httpheaders"
)

// Cookies represents a cookies manager.
type Cookies struct {
	baseURL *url.URL
	config  *Config
}

// cookieJar is a cookie jar that stores all cookies in memory
// and doesn't care about domains and paths.
type cookieJar struct {
	entries []*http.Cookie
	mu      sync.RWMutex
}

// New creates a new Cookies instance
func New(config *Config) (*Cookies, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	cookies := Cookies{config: config}

	if len(config.CookieBaseURL) > 0 {
		if u, err := url.Parse(config.CookieBaseURL); err == nil {
			cookies.baseURL = u
		} else {
			return nil, fmt.Errorf("can't parse cookie base URL: %w", err)
		}
	}

	return &cookies, nil
}

// SetCookies stores the cookies in the jar. For each source cookie it creates
// a new cookie with only Name, Value and Quoted fields set.
func (j *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	j.mu.Lock()
	defer j.mu.Unlock()

	// Remove all unimportant cookie params
	for _, c := range cookies {
		j.entries = append(j.entries, &http.Cookie{
			Name:   c.Name,
			Value:  c.Value,
			Quoted: c.Quoted,
		})
	}
}

// Cookies returns all stored cookies
func (j *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	j.mu.RLock()
	defer j.mu.RUnlock()

	// NOTE: do we need to clone, or we could just return a ref?
	return slices.Clone(j.entries)
}

// JarFromRequest creates a cookie jar from the given HTTP request
func (c *Cookies) JarFromRequest(r *http.Request) (jar http.CookieJar, err error) {
	// If cookie passthrough is disabled, return nil jar
	if !c.config.CookiePassthrough {
		return nil, nil
	}

	if c.config.CookiePassthroughAll {
		jar = &cookieJar{}
	} else {
		jar, err = cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			return nil, err
		}
	}

	if r == nil {
		return jar, nil
	}

	cookieBase := c.baseURL

	if !c.config.CookiePassthroughAll {
		if cookieBase == nil {
			scheme := r.Header.Get(httpheaders.XForwardedProto)
			if len(scheme) == 0 {
				scheme = "http"
			}

			host := r.Header.Get(httpheaders.XForwardedHost)
			if len(host) == 0 {
				host = r.Header.Get(httpheaders.Host)
			}
			if len(host) == 0 {
				host = r.Host
			}

			if len(host) == 0 {
				return jar, nil
			}

			port := r.Header.Get(httpheaders.XForwardedPort)
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
