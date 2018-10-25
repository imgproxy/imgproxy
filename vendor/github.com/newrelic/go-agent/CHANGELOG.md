## ChangeLog

## 2.1.0

* The Go Agent now supports distributed tracing.

  Distributed tracing lets you see the path that a request takes as it travels through your distributed system. By
  showing the distributed activity through a unified view, you can troubleshoot and understand a complex system better
  than ever before.

  Distributed tracing is available with an APM Pro or equivalent subscription. To see a complete distributed trace, you
  need to enable the feature on a set of neighboring services. Enabling distributed tracing changes the behavior of
  some New Relic features, so carefully consult the
  [transition guide](https://docs.newrelic.com/docs/transition-guide-distributed-tracing) before you enable this
  feature.

  To enable distributed tracing, set the following fields in your config.  Note that distributed tracing and cross
  application tracing cannot be used simultaneously.

```
  config := newrelic.NewConfig("Your Application Name", "__YOUR_NEW_RELIC_LICENSE_KEY__")
  config.CrossApplicationTracer.Enabled = false
  config.DistributedTracer.Enabled = true
```

  Please refer to the
  [distributed tracing section of the guide](GUIDE.md#distributed-tracing)
  for more detail on how to ensure you get the most out of the Go agent's distributed tracing support.

* Added functions [NewContext](https://godoc.org/github.com/newrelic/go-agent#NewContext)
  and [FromContext](https://godoc.org/github.com/newrelic/go-agent#FromContext)
  for adding and retrieving the Transaction from a Context.  Handlers
  instrumented by
  [WrapHandle](https://godoc.org/github.com/newrelic/go-agent#WrapHandle),
  [WrapHandleFunc](https://godoc.org/github.com/newrelic/go-agent#WrapHandleFunc),
  and [nrgorilla.InstrumentRoutes](https://godoc.org/github.com/newrelic/go-agent/_integrations/nrgorilla/v1#InstrumentRoutes)
  may use [FromContext](https://godoc.org/github.com/newrelic/go-agent#FromContext)
  on the request's context to access the Transaction.
  Thanks to @caarlos0 for the contribution!  Though [NewContext](https://godoc.org/github.com/newrelic/go-agent#NewContext)
  and [FromContext](https://godoc.org/github.com/newrelic/go-agent#FromContext)
  require Go 1.7+ (when [context](https://golang.org/pkg/context/) was added),
  [RequestWithTransactionContext](https://godoc.org/github.com/newrelic/go-agent#RequestWithTransactionContext) is always exported so that it can be used in all framework and library
  instrumentation.

## 2.0.0

* The `End()` functions defined on the `Segment`, `DatastoreSegment`, and
  `ExternalSegment` types now receive the segment as a pointer, rather than as
  a value. This prevents unexpected behaviour when a call to `End()` is
  deferred before one or more fields are changed on the segment.

  In practice, this is likely to only affect this pattern:

    ```go
    defer newrelic.DatastoreSegment{
      // ...
    }.End()
    ```

  Instead, you will now need to separate the literal from the deferred call:

    ```go
    ds := newrelic.DatastoreSegment{
      // ...
    }
    defer ds.End()
    ```

  When creating custom and external segments, we recommend using
  [`newrelic.StartSegment()`](https://godoc.org/github.com/newrelic/go-agent#StartSegment)
  and
  [`newrelic.StartExternalSegment()`](https://godoc.org/github.com/newrelic/go-agent#StartExternalSegment),
  respectively.

* Added GoDoc badge to README.  Thanks to @mrhwick for the contribution!

* `Config.UseTLS` configuration setting has been removed to increase security.
   TLS will now always be used in communication with New Relic Servers.

## 1.11.0

* We've closed the Issues tab on GitHub. Please visit our
  [support site](https://support.newrelic.com) to get timely help with any
  problems you're having, or to report issues.

* Added support for Cross Application Tracing (CAT). Please refer to the
  [CAT section of the guide](GUIDE.md#cross-application-tracing)
  for more detail on how to ensure you get the most out of the Go agent's new
  CAT support.

* The agent now collects additional metadata when running within Amazon Web
  Services, Google Cloud Platform, Microsoft Azure, and Pivotal Cloud Foundry.
  This information is used to provide an enhanced experience when the agent is
  deployed on those platforms.

## 1.10.0

* Added new `RecordCustomMetric` method to [Application](https://godoc.org/github.com/newrelic/go-agent#Application).
  This functionality can be used to track averages or counters without using
  custom events.
  * [Custom Metric Documentation](https://docs.newrelic.com/docs/agents/manage-apm-agents/agent-data/collect-custom-metrics)

* Fixed import needed for logrus.  The import Sirupsen/logrus had been renamed to sirupsen/logrus.
  Thanks to @alfred-landrum for spotting this.

* Added [ErrorAttributer](https://godoc.org/github.com/newrelic/go-agent#ErrorAttributer),
  an optional interface that can be implemented by errors provided to
  `Transaction.NoticeError` to attach additional attributes.  These attributes are
  subject to attribute configuration.

* Added [Error](https://godoc.org/github.com/newrelic/go-agent#Error), a type
  that allows direct control of error fields.  Example use:

```go
txn.NoticeError(newrelic.Error{
	// Message is returned by the Error() method.
	Message: "error message: something went very wrong",
	Class:   "errors are aggregated by class",
	Attributes: map[string]interface{}{
		"important_number": 97232,
		"relevant_string":  "zap",
	},
})
```

* Updated license to address scope of usage.

## 1.9.0

* Added support for [github.com/gin-gonic/gin](https://github.com/gin-gonic/gin)
  in the new `nrgin` package.
  * [Documentation](http://godoc.org/github.com/newrelic/go-agent/_integrations/nrgin/v1)
  * [Example](examples/_gin/main.go)

## 1.8.0

* Fixed incorrect metric rule application when the metric rule is flagged to
  terminate and matches but the name is unchanged.

* `Segment.End()`, `DatastoreSegment.End()`, and `ExternalSegment.End()` methods now return an
  error which may be helpful in diagnosing situations where segment data is unexpectedly missing.

## 1.7.0

* Added support for [gorilla/mux](http://github.com/gorilla/mux) in the new `nrgorilla`
  package.
  * [Documentation](http://godoc.org/github.com/newrelic/go-agent/_integrations/nrgorilla/v1)
  * [Example](examples/_gorilla/main.go)

## 1.6.0

* Added support for custom error messages and stack traces.  Errors provided
  to `Transaction.NoticeError` will now be checked to see if
  they implement [ErrorClasser](https://godoc.org/github.com/newrelic/go-agent#ErrorClasser)
  and/or [StackTracer](https://godoc.org/github.com/newrelic/go-agent#StackTracer).
  Thanks to @fgrosse for this proposal.

* Added support for [pkg/errors](https://github.com/pkg/errors).  Thanks to
  @fgrosse for this work.
  * [documentation](https://godoc.org/github.com/newrelic/go-agent/_integrations/nrpkgerrors)
  * [example](https://github.com/newrelic/go-agent/blob/master/_integrations/nrpkgerrors/nrpkgerrors.go)

* Fixed tests for Go 1.8.

## 1.5.0

* Added support for Windows.  Thanks to @ianomad and @lvxv for the contributions.

* The number of heap objects allocated is recorded in the
  `Memory/Heap/AllocatedObjects` metric.  This will soon be displayed on the "Go
  runtime" page.

* If the [DatastoreSegment](https://godoc.org/github.com/newrelic/go-agent#DatastoreSegment)
  fields `Host` and `PortPathOrID` are not provided, they will no longer appear
  as `"unknown"` in transaction traces and slow query traces.

* Stack traces will now be nicely aligned in the APM UI.

## 1.4.0

* Added support for slow query traces.  Slow datastore segments will now
 generate slow query traces viewable on the datastore tab.  These traces include
 a stack trace and help you to debug slow datastore activity.
 [Slow Query Documentation](https://docs.newrelic.com/docs/apm/applications-menu/monitoring/viewing-slow-query-details)

* Added new
[DatastoreSegment](https://godoc.org/github.com/newrelic/go-agent#DatastoreSegment)
fields `ParameterizedQuery`, `QueryParameters`, `Host`, `PortPathOrID`, and
`DatabaseName`.  These fields will be shown in transaction traces and in slow
query traces.

## 1.3.0

* Breaking Change: Added a timeout parameter to the `Application.Shutdown` method.

## 1.2.0

* Added support for instrumenting short-lived processes:
  * The new `Application.Shutdown` method allows applications to report
    data to New Relic without waiting a full minute.
  * The new `Application.WaitForConnection` method allows your process to
    defer instrumentation until the application is connected and ready to
    gather data.
  * Full documentation here: [application.go](application.go)
  * Example short-lived process: [examples/short-lived-process/main.go](examples/short-lived-process/main.go)

* Error metrics are no longer created when `ErrorCollector.Enabled = false`.

* Added support for [github.com/mgutz/logxi](github.com/mgutz/logxi).  See
  [_integrations/nrlogxi/v1/nrlogxi.go](_integrations/nrlogxi/v1/nrlogxi.go).

* Fixed bug where Transaction Trace thresholds based upon Apdex were not being
  applied to background transactions.

## 1.1.0

* Added support for Transaction Traces.

* Stack trace filenames have been shortened: Any thing preceding the first
  `/src/` is now removed.

## 1.0.0

* Removed `BetaToken` from the `Config` structure.

* Breaking Datastore Change:  `datastore` package contents moved to top level
  `newrelic` package.  `datastore.MySQL` has become `newrelic.DatastoreMySQL`.

* Breaking Attributes Change:  `attributes` package contents moved to top
  level `newrelic` package.  `attributes.ResponseCode` has become
  `newrelic.AttributeResponseCode`.  Some attribute name constants have been
  shortened.

* Added "runtime.NumCPU" to the environment tab.  Thanks sergeylanzman for the
  contribution.

* Prefixed the environment tab values "Compiler", "GOARCH", "GOOS", and
  "Version" with "runtime.".

## 0.8.0

* Breaking Segments API Changes:  The segments API has been rewritten with the
  goal of being easier to use and to avoid nil Transaction checks.  See:

  * [segments.go](segments.go)
  * [examples/server/main.go](examples/server/main.go)
  * [GUIDE.md#segments](GUIDE.md#segments)

* Updated LICENSE.txt with contribution information.

## 0.7.1

* Fixed a bug causing the `Config` to fail to serialize into JSON when the
  `Transport` field was populated.

## 0.7.0

* Eliminated `api`, `version`, and `log` packages.  `Version`, `Config`,
  `Application`, and `Transaction` now live in the top level `newrelic` package.
  If you imported the  `attributes` or `datastore` packages then you will need
  to remove `api` from the import path.

* Breaking Logging Changes

Logging is no longer controlled though a single global.  Instead, logging is
configured on a per-application basis with the new `Config.Logger` field.  The
logger is an interface described in [log.go](log.go).  See
[GUIDE.md#logging](GUIDE.md#logging).

## 0.6.1

* No longer create "GC/System/Pauses" metric if no GC pauses happened.

## 0.6.0

* Introduced beta token to support our beta program.

* Rename `Config.Development` to `Config.Enabled` (and change boolean
  direction).

* Fixed a bug where exclusive time could be incorrect if segments were not
  ended.

* Fix unit tests broken in 1.6.

* In `Config.Enabled = false` mode, the license must be the proper length or empty.

* Added runtime statistics for CPU/memory usage, garbage collection, and number
  of goroutines.

## 0.5.0

* Added segment timing methods to `Transaction`.  These methods must only be
  used in a single goroutine.

* The license length check will not be performed in `Development` mode.

* Rename `SetLogFile` to `SetFile` to reduce redundancy.

* Added `DebugEnabled` logging guard to reduce overhead.

* `Transaction` now implements an `Ignore` method which will prevent
  any of the transaction's data from being recorded.

* `Transaction` now implements a subset of the interfaces
  `http.CloseNotifier`, `http.Flusher`, `http.Hijacker`, and `io.ReaderFrom`
  to match the behavior of its wrapped `http.ResponseWriter`.

* Changed project name from `go-sdk` to `go-agent`.

## 0.4.0

* Queue time support added: if the inbound request contains an
`"X-Request-Start"` or `"X-Queue-Start"` header with a unix timestamp, the
agent will report queue time metrics.  Queue time will appear on the
application overview chart.  The timestamp may fractional seconds,
milliseconds, or microseconds: the agent will deduce the correct units.
