# 📑 Changelog (version/4 dev)

## ✨ 2025-08-27

### 🔄 Changed

- `If-None-Match` and `If-Modified-Since` are passed through to server request, `Etag` passed through response if `IMGPROXY_USE_ETAG` is true.
- `IMGPROXY_USE_ETAG` is now true by default.
- `IMGPROXY_USE_LAST_MODIFIED` is now true by default.

### ❌ Removed

- `Etag` calculations on the imgproxy side
