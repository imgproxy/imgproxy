# Health check

imgproxy comes with a built-in health check HTTP endpoint at `/health`.

`GET /health` returns HTTP Status `200 OK` if the server is started successfully.

You can use this for readiness/liveness probe when deploying with a container orchestration system such as Kubernetes.

## imgproxy health

imgproxy provides `imgproxy health` command that makes an HTTP request to the health endpoint based on `IMGPROXY_BIND` and `IMGPROXY_NETWORK` configs. It exits with `0` when the request is successful and with `1` otherwise. The command is handy to use with Docker Compose:

```yaml
healthcheck:
  test: [ "CMD", "imgproxy", "health" ]
  timeout: 10s
  interval: 10s
  retries: 3
```
