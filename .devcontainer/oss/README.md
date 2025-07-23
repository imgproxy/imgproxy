# .devcontainer setup for local development of `imgproxy`

All `imgproxy` dependencies are included in the `imgproxy-base` container image. Using this image for development is recommended.

If you want to develop locally without using Docker, please install: `vips`, `clang-format`,`lychee` and `lefthook`.

On MacOS:

```sh
brew install vips clang-format lychee lefthook
```

Then, run:
```sh
lefthook install
```

# Start the devcontainer

You can use [`air`](https://github.com/air-verse/air) for hot-reloading during development. Simply run: `air`.

Port `8080` is forwared to the host.

# Test images

[test images repo](https://github.com/imgproxy/test-images.git) will be automatically cloned or pulled to `.devcontainer/images` folder before the container starts.

[Try it](http://localhost:8080/insecure/rs:fit:300:200/plain/local:///kitten.jpg@png). -->
