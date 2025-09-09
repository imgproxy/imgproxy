# 📑 Changelog (version/4 dev)

## ✨ 2025-08-27

### 🔄 Changed

- `If-None-Match` is passed through to server request, `Etag` passed through from server response
if `IMGPROXY_USE_ETAG` is true.
- `IMGPROXY_USE_ETAG` is now true by default.
- `IMGPROXY_USE_LAST_MODIFIED` is now true by default.

### ❌ Removed

- `Etag` calculations on the imgproxy side

## 2025-09-09

### ❌ Removed

- `--keys`, `--salts` CLI args
