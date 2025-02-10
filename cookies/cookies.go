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

// PassthroughCookieJar is a custom CookieJar that can return either all cookies
// or only those according to the RFC6265 rules.
type PassthroughCookieJar struct {
	jar             http.CookieJar // underlying default jar (RFC6265 compliant)
	alwaysReturnAll bool           // when true, Cookies() returns all stored cookies
	mu              sync.Mutex     // protects allCookies and alwaysReturnAll
	allCookies      []*http.Cookie // holds every cookie added via SetCookies
}

// Default implementation doesn't expose cookies, so we need to keep track of them in our slice
func (ocj *PassthroughCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	ocj.jar.SetCookies(u, cookies)
	ocj.mu.Lock()
	// Simply append the new cookies.
	ocj.allCookies = append(ocj.allCookies, cookies...)
	ocj.mu.Unlock()
}

// Cookies returns cookies according to the runtime flag.
// If alwaysReturnAll is true, it returns all stored cookies.
// Otherwise, it returns the cookies for the given URL as computed by the underlying jar.
func (ocj *PassthroughCookieJar) Cookies(u *url.URL) []*http.Cookie {
	ocj.mu.Lock()
	defer ocj.mu.Unlock()
	if ocj.alwaysReturnAll {
		// Return a copy of all cookies.
		copied := make([]*http.Cookie, len(ocj.allCookies))
		copy(copied, ocj.allCookies)
		return copied
	}
	return ocj.jar.Cookies(u)
}

// SetAlwaysReturnAll lets you toggle the behavior at runtime.
func (ocj *PassthroughCookieJar) SetAlwaysReturnAll(always bool) {
	ocj.mu.Lock()
	ocj.alwaysReturnAll = always
	ocj.mu.Unlock()
}

func JarFromRequest(r *http.Request) (*PassthroughCookieJar, error) {
	orig_jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil, err
	}

	jar := &PassthroughCookieJar{
		jar:             orig_jar,
		alwaysReturnAll: config.CookieDisableChecks,
		allCookies:      make([]*http.Cookie, 0),
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
