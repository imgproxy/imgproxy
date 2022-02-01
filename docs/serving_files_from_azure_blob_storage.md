# Serving files from Azure Blob Storage

imgproxy can process images from Azure Blob Storage containers. To use this feature, do the following:

1. Set `IMGPROXY_USE_ABS` environment variable to `true`
2. Set `IMGPROXY_ABS_NAME` to your Azure account name and `IMGPROXY_ABS_KEY` to your Azure account key
4. _(optional)_ Specify the Azure Blob Storage endpoint with `IMGPROXY_ABS_ENDPOINT`
4. Use `abs://%bucket_name/%file_key` as the source image URL
