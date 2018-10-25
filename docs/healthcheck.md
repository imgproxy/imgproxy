# Health check

imgproxy comes with a built-in health check HTTP endpoint at `/health`.

`GET /health` returns HTTP Status `200 OK` if the server is started successfully.

You can use this for readiness/liveness probe when deploying with a container orchestration system such as Kubernetes.
