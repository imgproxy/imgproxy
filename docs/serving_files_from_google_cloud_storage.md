# Serving files from Google Cloud Storage

imgproxy can process images from Google Cloud Storage buckets. To use this feature, do the following:

1. Set `IMGPROXY_GCS_KEY` environment variable to the content of Google Cloud JSON key. Get more info about JSON keys: [https://cloud.google.com/iam/docs/creating-managing-service-account-keys](https://cloud.google.com/iam/docs/creating-managing-service-account-keys);
2. Use `gs://%bucket_name/%file_key` as the source image URL.

If you need to specify generation of the source object, you can use query string of the source URL:

```
gs://%bucket_name/%file_key?%generation
```
