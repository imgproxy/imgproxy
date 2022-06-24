# New Relic

imgproxy can send its metrics to New Relic. To use this feature, do the following:

1. Register at New Relic and get a license key.
2. Set the `IMGPROXY_NEW_RELIC_KEY` environment variable to the license key.
3. _(optional)_ Set the `IMGPROXY_NEW_RELIC_APP_NAME` environment variable to be the desired application name.
4. _(optional)_ Set the `IMGPROXY_NEW_RELIC_LABELS` environment variable to be the desired list of labels. Example: `label1=value1;label2=value2`.

imgproxy will send the following info to New Relic:

* CPU and memory usage
* Response time
* Image downloading time
* Image processing time
* Errors that occurred while downloading and processing an image
