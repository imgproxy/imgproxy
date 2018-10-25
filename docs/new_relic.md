# New Relic

imgproxy can send its metrics to New Relic. To use this feature, do the following:

1. Register at New Relic to get license key;
2. Set `IMGPROXY_NEW_RELIC_KEY` environment variable to the license key;
3. _(optional)_ Set `IMGPROXY_NEW_RELIC_APP_NAME` environment variable to the desired application name.

imgproxy will send the following info to New Relic:

* CPU and memory usage;
* Response time;
* Image downloading time;
* Image processing time;
* Errors that occurred while downloading and processing image.
