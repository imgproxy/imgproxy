# New Relic

imgproxy can send its metrics to New Relic. To use this feature, do the following:

1. Register at New Relic and get a license key.
2. Set the `IMGPROXY_NEW_RELIC_KEY` environment variable to the license key.
3. _(optional)_ Set the `IMGPROXY_NEW_RELIC_APP_NAME` environment variable to be the desired application name.
4. _(optional)_ Set the `IMGPROXY_NEW_RELIC_LABELS` environment variable to be the desired list of labels. Example: `label1=value1;label2=value2`.

imgproxy will send the following info to New Relic:

* CPU and memory usage
* Response time
* Queue time
* Image downloading time
* Image processing time
* Errors that occurred while downloading and processing an image

Additionally, imgproxy sends the following metrics over [Metrics API](https://docs.newrelic.com/docs/data-apis/ingest-apis/metric-api/introduction-metric-api/):

* `imgproxy.requests_in_progress`: the number of requests currently in progress
* `imgproxy.images_in_progress`: the number of images currently in progress
* `imgproxy.buffer.size`: a summary of the download/gzip buffers sizes (in bytes)
* `imgproxy.buffer.default_size`: calibrated default buffer size (in bytes)
* `imgproxy.buffer.max_size`: calibrated maximum buffer size (in bytes)
* `imgproxy.vips.memory`: libvips memory usage (in bytes)
* `imgproxy.vips.max_memory`: libvips maximum memory usage (in bytes)
* `imgproxy.vips.allocs`: the number of active vips allocations
