# Prometheus

imgproxy can collect its metrics for Prometheus. To use this feature, do the following:

1. Set `IMGPROXY_PROMETHEUS_BIND` environment variable. Note that you can't bind the main server and Prometheus to the same port;
2. Collect the metrics from any path on the specified binding.

imgproxy will collect the following metrics:

* `requests_total` - a counter of the total number of HTTP requests imgproxy processed;
* `errors_total` - a counter of the occurred errors separated by type (timeout, downloading, processing);
* `request_duration_seconds` - a histogram of the response latency (seconds);
* `download_duration_seconds` - a histogram of the source image downloading latency (seconds);
* `processing_duration_seconds` - a histogram of the image processing latency (seconds);
* `buffer_size_bytes` - a histogram of the download/gzip buffers sizes (bytes);
* `buffer_default_size_bytes` - calibrated default buffer size (bytes);
* `buffer_max_size_bytes` - calibrated maximum buffer size (bytes);
* Some useful Go metrics like memstats and goroutines count.
