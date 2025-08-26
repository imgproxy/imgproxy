# 📑 Changelog (version/4 dev)

## ✨ 2025-08-27

### 🔄 Changed

- `If-None-Match` and `If-Modified-Since` are passed through to server request, `Etag` passed through response if `IMGPROXY_USE_ETAG` is true.

### ❌ Removed

- `Etag` calculations on the imgproxy side
