# Serving files from Google Cloud Storage

imgproxy can process images from Google Cloud Storage buckets. To use this feature, do the following:

1. Set the `IMGPROXY_USE_GCS` environment variable to `true`.
2. [Set up credentials](#setup-credentials) to grant access to your bucket.
3. _(optional)_ Specify the Google Cloud Storage endpoint with `IMGPROXY_GCS_ENDPOINT`.
4. Use `gs://%bucket_name/%file_key` as the source image URL.

If you need to specify generation of the source object, you can use the query string of the source URL:

```
gs://%bucket_name/%file_key?%generation
```

### Setup credentials

If you run imgproxy inside Google Cloud infrastructure (Compute Engine, Kubernetes Engine, App Engine, and Cloud Functions, etc), and you have granted access to your bucket to the service account, you probably don't need to do anything here. imgproxy will try to use the credentials provided by Google.

Otherwise, set `IMGPROXY_GCS_KEY` environment variable to the content of Google Cloud JSON key. Get more info about JSON keys: [https://cloud.google.com/iam/docs/creating-managing-service-account-keys](https://cloud.google.com/iam/docs/creating-managing-service-account-keys).
