# Honeybadger for Go

[![Build Status](https://travis-ci.org/honeybadger-io/honeybadger-go.svg?branch=master)](https://travis-ci.org/honeybadger-io/honeybadger-go)

Go (golang) support for the :zap: [Honeybadger error
notifier](https://www.honeybadger.io/). Receive instant notification of panics
and errors in your Go applications.

## Getting Started


### 1. Install the library

To install, grab the package from GitHub:

```sh
go get github.com/honeybadger-io/honeybadger-go
```

Then add an import to your application code:

```go
import "github.com/honeybadger-io/honeybadger-go"
```

### 2. Set your API key

Finally, configure your API key:

```go
honeybadger.Configure(honeybadger.Configuration{APIKey: "your api key"})
```

You can also configure Honeybadger via environment variables. See
[Configuration](#configuration) for more information.

### 3. Enable automatic panic reporting

#### Panics during HTTP requests

To automatically report panics which happen during an HTTP request, wrap your
`http.Handler` function with `honeybadger.Handler`:

```go
log.Fatal(http.ListenAndServe(":8080", honeybadger.Handler(handler)))
```

Request data such as cookies and params will automatically be reported with
errors which happen inside `honeybadger.Handler`. Make sure you recover from
panics after honeybadger's Handler has been executed to ensure all panics are
reported.

#### Unhandled Panics


To report all unhandled panics which happen in your application
the following can be added to `main()`:

```go
func main() {
  defer honeybadger.Monitor()
  // application code...
}
```

#### Manually Reporting Errors

To report an error manually, use `honeybadger.Notify`:

```go
if err != nil {
  honeybadger.Notify(err)
}
```


## Sample Application

If you'd like to see the library in action before you integrate it with your apps, check out our [sample application](https://github.com/honeybadger-io/crywolf-go). 

You can deploy the sample app to your Heroku account by clicking this button:

[![Deploy](https://www.herokucdn.com/deploy/button.png)](https://heroku.com/deploy?template=https://github.com/honeybadger-io/crywolf-go)

Don't forget to destroy the Heroku app after you're done so that you aren't charged for usage.

The code for the sample app is [available on Github](https://github.com/honeybadger-io/crywolf-go), in case you'd like to read through it, or run it locally.


## Configuration

To set configuration options, use the `honeybadger.Configuration` method, like so:

```go
honeybadger.Configure(honeybadger.Configuration{
  APIKey: "your api key", 
  Env: "staging"
})
```
The following options are available to you:

|  Name | Type | Default | Example | Environment variable |
| ----- | ---- | ------- | ------- | -------------------- |
| APIKey | `string` | `""` | `"badger01"` | `HONEYBADGER_API_KEY` |
| Root | `string` | The current working directory | `"/path/to/project"` | `HONEYBADGER_ROOT` |
| Env | `string` | `""` | `"production"` | `HONEYBADGER_ENV` |
| Hostname | `string` | The hostname of the current server. | `"badger01"` | `HONEYBADGER_HOSTNAME` |
| Endpoint | `string` | `"https://api.honeybadger.io"` | `"https://honeybadger.example.com/"` | `HONEYBADGER_ENDPOINT` |
| Timeout | `time.Duration` | 3 seconds | `10 * time.Second` | `HONEYBADGER_TIMEOUT` (nanoseconds) |
| Logger | `honeybadger.Logger` | Logs to stderr | `CustomLogger{}` | n/a |
| Backend | `honeybadger.Backend` | HTTP backend | `CustomBackend{}` | n/a |


## Public Interface

### `honeybadger.Notify()`: Send an error to Honeybadger.

If you've handled a panic in your code, but would still like to report the error to Honeybadger, this is the method for you. 

#### Examples:

```go
if err != nil {
  honeybadger.Notify(err)
}
```

You can also add local context using an optional second argument when calling
`honeybadger.Notify`:

```go
honeybadger.Notify(err, honeybadger.Context{"user_id": 2})
```

Honeybadger uses the error's class name to group similar errors together. If
your error classes are often generic (such as `errors.errorString`), you can
improve grouping by overriding the default with something more unique:

```go
honeybadger.Notify(err, honeybadger.ErrorClass{"CustomClassName"})
```

To override grouping entirely, you can send a custom fingerprint. All errors
with the same fingerprint will be grouped together:

```go
honeybadger.Notify(err, honeybadger.Fingerprint{"A unique string"})
```

To tag errors in Honeybadger:

```go
honeybadger.Notify(err, honeybadger.Tags{"timeout", "http"})
```

---


### `honeybadger.SetContext()`: Set metadata to be sent if an error occurs

This method lets you set context data that will be sent if an error should occur.

For example, it's often useful to record the current user's ID when an error occurs in a web app. To do that, just use `SetContext` to set the user id on each request. If an error occurs, the id will be reported with it.

**Note**: This method is currently shared across goroutines, and therefore may not be optimal for use in highly concurrent use cases, such as HTTP requests. See [issue #35](https://github.com/honeybadger-io/honeybadger-go/issues/35).

#### Examples:

```go
honeybadger.SetContext(honeybadger.Context{
  "user_id": 1,
})
```

---

### ``defer honeybadger.Monitor()``: Automatically report panics from your functions

To automatically report panics in your functions or methods, add
`defer honeybadger.Monitor()` to the beginning of the function or method you wish to monitor.
 

#### Examples:

```go
func risky() {
  defer honeybadger.Monitor()
  // risky business logic...
}
```

__Important:__ `honeybadger.Monitor()` will re-panic after it reports the error, so make sure that it is only called once before recovering from the panic (or allowing the process to crash).

---

### ``honeybadger.BeforeNotify()``: Add a callback to skip or modify error notification.

Sometimes you may want to modify the data sent to Honeybadger right before an
error notification is sent, or skip the notification entirely. To do so, add a
callback using `honeybadger.BeforeNotify()`.

#### Examples:

```go
honeybadger.BeforeNotify(
  func(notice *honeybadger.Notice) error {
    if notice.ErrorClass == "SkippedError" {
      return fmt.Errorf("Skipping this notification")
    }
    // Return nil to send notification for all other classes.
    return nil
  }
)
```

To modify information:

```go
honeybadger.BeforeNotify(
  func(notice *honeybadger.Notice) error {
    // Errors in Honeybadger will always have the class name "GenericError".
    notice.ErrorClass = "GenericError"
    return nil
  }
)
```

---

### ``honeybadger.NewNullBackend()``: Disable data reporting.

`NewNullBackend` creates a backend which swallows all errors and does not send them to Honeybadger. This is useful for development and testing to disable sending unnecessary errors.

#### Examples:

```go
honeybadger.Configure(honeybadger.Configuration{Backend: honeybadger.NewNullBackend()})
```

---

## Creating a new client

In the same way that the log library provides a predefined "standard" logger, honeybadger defines a standard client which may be accessed directly via `honeybadger`. A new client may also be created by calling `honeybadger.New`:

```go
hb := honeybadger.New(honeybadger.Configuration{APIKey: "some other api key"})
hb.Notify("This error was reported by an alternate client.")
```

## Grouping

Honeybadger groups by the error class and the first line of the backtrace by
default. In some cases it may be desirable to  provide your own grouping
algorithm. One use case for this is `errors.errorString`. Because that type is
used for many different types of errors in Go, Honeybadger will appear to group
unrelated errors together. Here's an example of providing a custom fingerprint
which will group `errors.errorString` by message instead:

```go
honeybadger.BeforeNotify(
  func(notice *honeybadger.Notice) error {
    if notice.ErrorClass == "errors.errorString" {
      notice.Fingerprint = notice.Message
    }
    return nil
  }
)
```

Note that in this example, the backtrace is ignored. If you want to group by
message *and* backtrace, you could append data from `notice.Backtrace` to the
fingerprint string.

An alternate approach would be to override `notice.ErrorClass` with a more
specific class name that may be inferred from the message.

## Versioning

We use [Semantic Versioning](http://semver.org/) to version releases of
honeybadger-go. Because there is no official method to specify version
dependencies in Go, we will do our best never to introduce a breaking change on
the master branch of this repo after reaching version 1. Until we reach version
1 there is a small chance that we may introduce a breaking change (changing the
signature of a function or method, for example), but we'll always tag a new
minor release and broadcast that we made the change.

If you're concerned about versioning, there are two options:

### Vendor your dependencies

If you're really concerned about changes to this library, then copy it into your
source control management system so that you can perform upgrades on your own
time.

### Use gopkg.in

Rather than importing directly from GitHub, [gopkg.in](http://gopkg.in/) allows
you to use their special URL format to transparently import a branch or tag from
GitHub. Because we tag each release, using gopkg.in can enable you to depend
explicitly on a certain version of this library. Importing from gopkg.in instead
of directly from GitHub is as easy as:

```go
import "gopkg.in/honeybadger-io/honeybadger-go.v0"
```

Check out the [gopkg.in](http://gopkg.in/) homepage for more information on how
to request versions.

## Changelog

See https://github.com/honeybadger-io/honeybadger-go/blob/master/CHANGELOG.md

## Contributing

If you're adding a new feature, please [submit an issue](https://github.com/honeybadger-io/honeybadger-go/issues/new) as a preliminary step; that way you can be (moderately) sure that your pull request will be accepted.

### To contribute your code:

1. Fork it.
2. Create a topic branch `git checkout -b my_branch`
3. Commit your changes `git commit -am "Boom"`
3. Push to your branch `git push origin my_branch`
4. Send a [pull request](https://github.com/honeybadger-io/honeybadger-go/pulls)

### License

This library is MIT licensed. See the [LICENSE](https://raw.github.com/honeybadger-io/honeybadger-go/master/LICENSE) file in this repository for details.
