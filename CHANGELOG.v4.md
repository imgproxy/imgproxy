# 📑 Changelog (version/4 dev)

## 2026-03-04

### 🔄 Changed

- Changed SVG rendering DPI to 96 (from 72) to match W3C recommendations.

## 2026-02-03

### 🔄 Changed

- Fixed flatten behaviour. Black and white source images are now converted to color if background is non-grey.

## 2026-01-09

### 🆕 Added

- `IMGPROXY_PRESERVE_HDR`. BPP/colorspace of source image is kept when possible.

## 2025-12-12

### ❌ Removed

- `--presets` CLI argument in favor of `IMGPROXY_PRESETS_PATH`.

## 2025-12-09

### 🔄 Changed

- [Pro] Improved SVG minification compression (>x2) and speed.

## 2025-12-02

### 🔄 Changed

- [Pro] SVG injected styles are wrapped in <![CDATA[]]>

## 2025-11-05

### 🆕 Added

- [Pro] Introduced local cache. Please, check the docs for the available configuration options.

## 2025-10-31

### ❌ Removed

- Deprecated `download_duration_seconds` and `processing_duration_seconds` histograms from Prometheus metrics.

### 🆕 Added

- Propagation of tracing headers to external requests. Introduced `IMGPROXY_(NEW_RELIC|DATADOG|OPEN_TELEMETRY)_PROPAGATE_EXTERNAL` env var to control this behavior.

## 2025-10-29

- Introduced `IMGPROXY_(ABS|GCS|S3|SWIFT)_(ALLOWED|DENIED)_BUCKETS` env var

## 2025-10-20

### 🆕 Added

- [pro] Perceptual Hash (pHash) calculation

## 2025-10-02

### 🆕 Added

- Custom metrics are now reported as timeslices (see [New Relic’s documentation](https://docs.newrelic.com/docs/apm/agents/manage-apm-agents/agent-data/collect-custom-metrics/) for details). Metric names have been changed from "imgproxy.X" to "Custom/imgproxy/X".

## 2025-10-01

### 🆕 Added

- `DD_TRACE_AGENT_PORT` (default: 8126) as a default DataDog trace agent port.

## 2025-09-30

### ❌ Removed

- Deprecated `IMGPROXY_OPEN_TELEMETRY_ENDPOINT` is removed.
- Deprecated `IMGPROXY_OPEN_TELEMETRY_PROTOCOL` is removed.
- Deprecated `IMGPROXY_OPEN_TELEMETRY_GRPC_INSECURE` is removed.
- Deprecated `IMGPROXY_OPEN_TELEMETRY_SERVICE_NAME` is removed.
- Deprecated `IMGPROXY_OPEN_TELEMETRY_PROPAGATORS` is removed.
- Deprecated `IMGPROXY_OPEN_TELEMETRY_CONNECTION_TIMEOUT` is removed.

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

## 2025-09-16

### ❌ Removed

- `IMGPROXY_UNSHARPENING_MODE`, `IMGPROXY_UNSHARPENING_WEIGHT`, `IMGPROXY_UNSHARPENING_DIVIDER` configs. Use `IMGPROXY_UNSHARP_MASKING_MODE`, `IMGPROXY_UNSHARP_MASKING_WEIGHT`, `IMGPROXY_UNSHARP_MASKING_DIVIDER` instead.

## 2025-09-18

### ❌ Removed

- `gif_options` processing option

## 2025-09-11

### ❌ Removed

- Deprecated `IMGPROXY_CONCURRENCY` removed.

## 2025-09-09

### ❌ Removed

- `--keys`, `--salts` CLI args

## ✨ 2025-08-27

### 🔄 Changed

- `If-None-Match` is passed through to server request, `Etag` passed through from server response
if `IMGPROXY_USE_ETAG` is true.
- `IMGPROXY_USE_ETAG` is now true by default.
- `IMGPROXY_USE_LAST_MODIFIED` is now true by default.

### ❌ Removed

- `Etag` calculations on the imgproxy side
