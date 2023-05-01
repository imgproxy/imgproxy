# OpenTelemetry

imgproxy can send request traces to an OpenTelemetry collector. To use this feature, do the following:

1. Install & configure the [OpenTelemetry collector](https://opentelemetry.io/docs/collector/).
2. Specify the collector endpoint (`host:port`) with `IMGPROXY_OPEN_TELEMETRY_ENDPOINT` and the collector protocol with `IMGPROXY_OPEN_TELEMETRY_PROTOCOL`. Supported protocols are:
    * `grpc` _(default)_
    * `https`
    * `http`.
3. _(optional)_ Set the `IMGPROXY_OPEN_TELEMETRY_SERVICE_NAME` environment variable to be the desired service name.
4. _(optional)_ Set the `IMGPROXY_OPEN_TELEMETRY_PROPAGATORS` environment variable to be the desired list of text map propagators. Supported propagators are:
    * `tracecontext`: [W3C Trace Context](https://www.w3.org/TR/trace-context/)
    * `baggage`: [W3C Baggage](https://www.w3.org/TR/baggage/)
    * `b3`: [B3 Single](https://opentelemetry.io/docs/reference/specification/context/api-propagators/#configuration)
    * `b3multi`: [B3 Multi](https://opentelemetry.io/docs/reference/specification/context/api-propagators/#configuration)
    * `jaeger`: [Jaeger](https://www.jaegertracing.io/docs/1.21/client-libraries/#propagation-format)
    * `xray`: [AWS X-Ray](https://docs.aws.amazon.com/xray/latest/devguide/xray-concepts.html#xray-concepts-tracingheader)
    * `ottrace`: [OT Trace](https://github.com/opentracing?q=basic&type=&language=)
5. _(optional)_ [Set up TLS certificates](#tls-configuration) or set `IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE` to `false` to use secure connection without TLS certificates set.
6. _(optional)_ Set `IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS` to `true` to enable sending metrics via OpenTelemetry Metrics API.
7. _(optional)_ Set `IMGPROXY_OPEN_TELEMETRY_TRACE_ID_GENERATOR` to environment variable to be the desired trace ID generator. Supported values are:
    * `xray`: _(default)_ Amazon X-Ray compatible trace ID generator
    * `random`: random trace ID generator

imgproxy will send the following info to the collector:

* Response time
* Queue time
* Image downloading time
* Image processing time
* Errors that occurred while downloading and processing an image

If `IMGPROXY_OPEN_TELEMETRY_ENABLE_METRICS` is set to `true`, imgproxy will also send the following metrics to the collector:

* `requests_in_progress`: the number of requests currently in progress
* `images_in_progress`: the number of images currently in progress
* `buffer_size_bytes`: a histogram of buffer sizes (in bytes)
* `buffer_default_size_bytes`: calibrated default buffer size (in bytes)
* `buffer_max_size_bytes`: calibrated maximum buffer size (in bytes)
* `vips_memory_bytes`: libvips memory usage
* `vips_max_memory_bytes`: libvips maximum memory usage
* `vips_allocs`: the number of active vips allocations
* Some useful Go metrics like memstats and goroutines count

## TLS Configuration

If your OpenTelemetry collector is secured with TLS, you may need to specify the collector's certificate on the imgproxy side:

* `IMGPROXY_OPEN_TELEMETRY_SERVER_CERT`: OpenTelemetry collector TLS certificate, PEM-encoded (you can replace line breaks with `\n`). Default: blank

If your collector uses mTLS for mutual authentication, you'll also need to specify the client's certificate/key pair:

* `IMGPROXY_OPEN_TELEMETRY_CLIENT_CERT`: OpenTelemetry client TLS certificate, PEM-encoded (you can replace line breaks with `\n`). Default: blank
* `IMGPROXY_OPEN_TELEMETRY_CLIENT_KEY`: OpenTelemetry client TLS key, PEM-encoded (you can replace line breaks with `\n`). Default: blank
