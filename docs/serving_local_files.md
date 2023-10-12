# Serving local files

imgproxy can be configured to process files from your local filesystem. To use this feature, do the following:

1. Set `IMGPROXY_LOCAL_FILESYSTEM_ROOT` environment variable to your local images directory path.
2. Use `local:///path/to/image.jpg` as the source image URL.

### Example

Assume you want to process an image that is stored locally at `/path/to/project/images/logos/evil_martians.png`. Run imgproxy with `IMGPROXY_LOCAL_FILESYSTEM_ROOT` set to your images directory:

```bash
IMGPROXY_LOCAL_FILESYSTEM_ROOT=/path/to/project/images imgproxy
```

Then, use the path inside this directory as the source URL:

```
local:///logos/evil_martians.png
```

The URL for resizing this image to fit 300x200 will look like this:

```
http://imgproxy.example.com/insecure/rs:fit:300:200:no:0/plain/local:///logos/evil_martians.png@jpg
```
