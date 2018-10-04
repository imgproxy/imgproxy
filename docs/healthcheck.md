# Health check

There is a special endpoint `/health`, which returns HTTP Status `200 OK` after the server successfully starts. This can be used for readiness/liveness probe in your containers system such as Kubernetes.
