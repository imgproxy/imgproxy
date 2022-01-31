# Health check

imgproxy comes with a built-in health check HTTP endpoint at `/health`.

`GET /health` returns an HTTP Status of `200 OK` if the server has been successfully started.

You can use this for readiness/liveness probes when deploying with a container orchestration system such as Kubernetes.

## imgproxy health

imgproxy provides an `imgproxy health` command that makes an HTTP request to the health endpoint based on the `IMGPROXY_BIND` and `IMGPROXY_NETWORK` configs. It exits with `0` when the request is successful and with `1` otherwise. The command is handy to use with Docker Compose:

```yaml
healthcheck:
  test: [ "CMD", "imgproxy", "health" ]
  timeout: 10s
  interval: 10s
  retries: 3
```
