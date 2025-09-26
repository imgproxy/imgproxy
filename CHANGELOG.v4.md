# 📑 Changelog (version/4 dev)

## 2025-09-26

### ❌ Removed

- Deprecated `IMGPROXY_WRITE_TIMEOUT` is removed.
- Deprecated `IMGPROXY_READ_TIMEOUT` is removed.
- Obsolete `IMGPROXY_MAX_SVG_CHECK_BYTES` is removed.
- Obsolete `IMGPROXY_ETAG_BUSTER` is removed.
- `IMGPROXY_USE_*` behaviour changed: now, it does not rely on the key

## 2025-09-25

### 🔄 Changed

- `IMGPROXY_USE_GCS` is not automatically set if gcs key is present anymore.

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
