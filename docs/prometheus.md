# Prometheus

imgproxy can collect metrics for Prometheus. To use this feature, do the following:

1. Set the `IMGPROXY_PROMETHEUS_BIND` environment variable to the address and port that will be listened to by the Prometheus server. Note that you can't bind the main server and Prometheus to the same port.
2. _(optional)_ Set the `IMGPROXY_PROMETHEUS_NAMESPACE` to prepend prefix to the names of metrics, i.e. with `IMGPROXY_PROMETHEUS_NAMESPACE=imgproxy` names will appear like `imgproxy_requests_total`.
3. Collect the metrics from any path on the specified binding.

imgproxy will collect the following metrics:

* `requests_total`: a counter with the total number of HTTP requests imgproxy has processed
* `errors_total`: a counter of the occurred errors separated by type (timeout, downloading, processing)
* `request_duration_seconds`: a histogram of the request latency (in seconds)
* `request_span_duration_seconds`: a histogram of the request latency (in seconds) separated by span (queue, downloading, processing)
* `requests_in_progress`: the number of requests currently in progress
* `images_in_progress`: the number of images currently in progress
* `buffer_size_bytes`: a histogram of the download/gzip buffers sizes (in bytes)
* `buffer_default_size_bytes`: calibrated default buffer size (in bytes)
* `buffer_max_size_bytes`: calibrated maximum buffer size (in bytes)
* `vips_memory_bytes`: libvips memory usage
* `vips_max_memory_bytes`: libvips maximum memory usage
* `vips_allocs`: the number of active vips allocations
* Some useful Go metrics like memstats and goroutines count

### Deprecated metrics

The following metrics are deprecated and can be removed in future versions. Use `request_span_duration_seconds` instead.

* `download_duration_seconds`: a histogram of the source image downloading latency (in seconds)
* `processing_duration_seconds`: a histogram of the image processing latency (in seconds)

