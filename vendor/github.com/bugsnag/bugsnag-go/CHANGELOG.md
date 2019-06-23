# Changelog

## 1.4.0 (2018-11-19)

This release is a big non-breaking revamp of the notifier. Most importantly, this release introduces session tracking to Go applications.

As of this release we require that you use Go 1.7 or higher.

### Features

* Session tracking to be able to show a stability score in the dashboard. Automatic recording of sessions for net/http, gin, revel, negroni and martini. Automatic capturing of sessions can be disabled using the `AutoCaptureSessions` configuration parameter.
* Automatic recording of HTTP request information such as HTTP method, headers, URL and query parameters.

### Enhancements

* Migrate report payload version from 3 to 4.
* Improve test coverage and introduce maze runner tests. Simplify integration tests for Negroni, Gin and Martini.
* Deprecate the use of the old `Endpoint` configuration parameter, and allow users of on-premise to configure both the notify endpoint and the sessions endpoint.
* `bugsnag.Notify()` now accepts a `context.Context` object, generally from `*http.Request`'s `r.Context()`, which Bugsnag can extract session and request information from.
* Improve and augment examples (`bugsnag_example_test.go`) for documentation.
* Improve example applications (`examples/` directory) to get up and running faster.
* Clarify and improve GoDocs.
* Improved serialization performance and safety of the report payload.
* Filter HTTP headers based on the `FiltersParams`.
* Revel enhancements:
    * Ensure all non-code configuration options are configurable from config file.
    * Stop using deprecated logger.
    * Attempt to configure a what we can from the revel configuration options.
* Make NotifyReleaseStages work consistently with other notifiers, both for sessions and for reports.
* Also filter out 'authorization' and 'cookie' by default, to match other notifiers.

### Bug fixes

* Address compile errors test failures that failed the build.
* Don't crash when calling `bugsnag.Notify(nil)`
* Other minor bug fixes that came to light after improving test coverage.

## 1.3.2 (2018-10-05)

### Bug fixes

* Ensure error reports for fatal crashes gets sent
  [#77](https://github.com/bugsnag/bugsnag-go/pull/77)

## 1.3.1 (2018-03-14)

### Bug fixes

* Add support for Revel v0.18
  [#63](https://github.com/bugsnag/bugsnag-go/pull/63)
  [Cameron Halter](https://github.com/EightB1ts)

## 1.3.0 (2017-10-02)

### Enhancements

* Track whether an error report was captured automatically
* Add SourceRoot as a configuration option, defaulting to `$GOPATH`

## 1.2.2 (2017-08-25)

### Bug fixes

* Point osext dependency at upstream, update with fixes

## 1.2.1 (2017-07-31)

### Bug fixes

* Improve goroutine panic reporting by sending reports synchronously in the
  case that a goroutine is about to be cleaned up
  [#52](https://github.com/bugsnag/bugsnag-go/pull/52)

## 1.2.0 (2017-07-03)

### Enhancements

* Support custom stack frame implementations
  [alexanderwilling](https://github.com/alexanderwilling)
  [#43](https://github.com/bugsnag/bugsnag-go/issues/43)

* Support app.type in error reports
  [Jascha Ephraim](https://github.com/jaschaephraim)
  [#51](https://github.com/bugsnag/bugsnag-go/pull/51)

### Bug fixes

* Mend nil pointer panic in metadata
  [Johan Sageryd](https://github.com/jsageryd)
  [#46](https://github.com/bugsnag/bugsnag-go/pull/46)

## 1.1.1 (2016-12-16)

### Bug fixes

* Replace empty error class property in reports with "error"

## 1.1.0 (2016-11-07)

### Enhancements

* Add middleware for Gin
  [Mike Bull](https://github.com/bullmo)
  [#40](https://github.com/bugsnag/bugsnag-go/pull/40)

* Add middleware for Negroni
  [am-manideep](https://github.com/am-manideep)
  [#28](https://github.com/bugsnag/bugsnag-go/pull/28)

* Support stripping subpackage names
  [Facundo Ferrer](https://github.com/fjferrer)
  [#25](https://github.com/bugsnag/bugsnag-go/pull/25)

* Support using `ErrorWithCallers` to create a stacktrace for errors
  [Conrad Irwin](https://github.com/ConradIrwin)
  [#35](https://github.com/bugsnag/bugsnag-go/pull/35)

## 1.0.5

### Bug fixes

* Avoid swallowing errors which occur upon delivery

1.0.4
-----

- Fix appengine integration broken by 1.0.3

1.0.3
-----

- Allow any Logger with a Printf method.

1.0.2
-----

- Use bugsnag copies of dependencies to avoid potential link rot

1.0.1
-----

- gofmt/golint/govet docs improvements.

1.0.0
-----
