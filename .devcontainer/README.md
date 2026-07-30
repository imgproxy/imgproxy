# .devcontainer setup for local development of `imgproxy`

All `imgproxy` dependencies are included in the `imgproxy-base` container image. Using this image for development is recommended.

If you want to develop locally without using Docker, please install: `vips`, `clang-format` and `lychee`.

On MacOS:

```sh
brew install vips clang-format lychee
```

If you'd rather use the devcontainer/Docker path, you'll need Docker and Go on your
host machine. Go is required because `guard_docker` (see `.runrc`) shells out to
`go tool gojq` to read `devcontainer.json` when re-invoking a `./run` task inside the
base container.

Then, run:
```sh
lefthook install
```

# Start the devcontainer

You can use [`air`](https://github.com/air-verse/air) for hot-reloading during development. Simply run: `go tool air`.

Port `8081` is forwared to the host.

# Test images

[test images repo](https://github.com/imgproxy/test-images.git) will be automatically cloned or pulled to `.devcontainer/images` folder before the container starts.

Use `./run devcontainer` to attach to the running devcontainer instance.

[Try it](http://localhost:8081/insecure/rs:fit:300:200/plain/local:///kitten.jpg@png). -->
