# Datadog

imgproxy can send its metrics to Datadog. To use this feature, do the following:

1. Install & configure the Datadog Trace Agent (>= 5.21.1).
2. Set the `IMGPROXY_DATADOG_ENABLE` environment variable to `true`.
3. Configure the Datadog tracer using `ENV` variables provided by [the package](https://github.com/DataDog/dd-trace-go):

    * `DD_AGENT_HOST`: sets the address to connect to for sending metrics to the Datadog Agent. Default: `localhost`
    * `DD_TRACE_AGENT_PORT`: sets the Datadog Agent Trace port. Default: `8126`
    * `DD_DOGSTATSD_PORT`: set the DogStatsD port. Default: `8125`
    * `DD_SERVICE`: sets the desired application name. Default: `imgproxy`
    * `DD_ENV`: specifies the environment to which all traces will be submitted. Default: empty
    * `DD_TRACE_SOURCE_HOSTNAME`: specifies the hostname with which to mark outgoing traces. Default: empty
    * `DD_TRACE_REPORT_HOSTNAME`: when `true`, sets hostname to `os.Hostname()` with which to mark outgoing traces. Default: `false`
    * `DD_TAGS`: sets a key/value pair which will be set as a tag on all traces. Example: `DD_TAGS=datacenter:njc,key2:value2`. Default: empty
    * `DD_TRACE_ANALYTICS_ENABLED`: allows specifying whether Trace Search & Analytics should be enabled for integrations. Default: `false`
    * `DD_RUNTIME_METRICS_ENABLED`: enables automatic collection of runtime metrics every 10 seconds. Default: `false`
    * `DD_TRACE_STARTUP_LOGS`: causes various startup info to be written when the tracer starts. Default: `true`
    * `DD_TRACE_DEBUG`: enables detailed logs. Default: `false`
4. _(optional)_ Set the `IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS` environment variable to `true` to collect the [additional metrics](#additional-metrics).

imgproxy will send the following info to Datadog:

* Response time
* Queue time
* Image downloading time
* Image processing time
* Errors that occurred while downloading and processing image

## Additional metrics

When the `IMGPROXY_DATADOG_ENABLE_ADDITIONAL_METRICS` environment variable is set to `true`, imgproxy will send the following additional metrics to Datadog:

* `imgproxy.requests_in_progress`: the number of requests currently in progress
* `imgproxy.images_in_progress`: the number of images currently in progress
* `imgproxy.buffer.size`: a histogram of the download/gzip buffers sizes (in bytes)
* `imgproxy.buffer.default_size`: calibrated default buffer size (in bytes)
* `imgproxy.buffer.max_size`: calibrated maximum buffer size (in bytes)
* `imgproxy.vips.memory`: libvips memory usage (in bytes)
* `imgproxy.vips.max_memory`: libvips maximum memory usage (in bytes)
* `imgproxy.vips.allocs`: the number of active vips allocations
