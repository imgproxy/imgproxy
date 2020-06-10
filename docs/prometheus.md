# Prometheus

imgproxy can collect its metrics for Prometheus. To use this feature, do the following:

1. Set `IMGPROXY_PROMETHEUS_BIND` environment variable. Note that you can't bind the main server and Prometheus to the same port;
2. _(optional)_ Set `IMGPROXY_PROMETHEUS_NAMESPACE` to prepend prefix to the names of metrics.
   I.e. with `IMGPROXY_PROMETHEUS_NAMESPACE=imgproxy` names will look like `imgproxy_requests_total`.
3. Collect the metrics from any path on the specified binding.

imgproxy will collect the following metrics:

* `requests_total` - a counter of the total number of HTTP requests imgproxy processed;
* `errors_total` - a counter of the occurred errors separated by type (timeout, downloading, processing);
* `request_duration_seconds` - a histogram of the response latency (seconds);
* `download_duration_seconds` - a histogram of the source image downloading latency (seconds);
* `processing_duration_seconds` - a histogram of the image processing latency (seconds);
* `buffer_size_bytes` - a histogram of the download/gzip buffers sizes (bytes);
* `buffer_default_size_bytes` - calibrated default buffer size (bytes);
* `buffer_max_size_bytes` - calibrated maximum buffer size (bytes);
* `vips_memory_bytes` - libvips memory usage;
* `vips_max_memory_bytes` - libvips maximum memory usage;
* `vips_allocs` - the number of active vips allocations;
* Some useful Go metrics like memstats and goroutines count.
