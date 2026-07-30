# .devcontainer setup for local development of `imgproxy`

All `imgproxy` dependencies are included in the `imgproxy-base` container image. Using this image for development is recommended.

You'll need Docker (with the Compose plugin, i.e. `docker compose`) on your host machine. `guard_docker`
(see `.runrc`) uses `docker compose run` against
`.devcontainer/docker-compose.yml` to re-invoke a `./run` task inside the
base container.

# Install git hooks

Run:
```sh
go tool lefthook install
```

# Start the devcontainer

You can use [`air`](https://github.com/air-verse/air) for hot-reloading during development. Simply run: `go tool air`.

Port `8081` is forwared to the host.

# Test images

[test images repo](https://github.com/imgproxy/test-images.git) will be automatically cloned or pulled to `.devcontainer/images` folder before the container starts.

Use `./run devcontainer` to attach to the running devcontainer instance.
