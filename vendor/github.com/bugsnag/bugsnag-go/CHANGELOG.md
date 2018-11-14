# Changelog

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
