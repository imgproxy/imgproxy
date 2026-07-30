# Contributing to imgproxy

Thanks for your interest in contributing! This document covers how to set up a local
dev environment and the day-to-day commands you'll use.

## Prerequisites

imgproxy's dev workflow is built around `./run`, a small task dispatcher
([run.sh](https://github.com/imgproxy/run.sh), a Makefile replacement — see `./run`,
`.runrc`, and `bin/*.sh`) that most tasks run inside the project's pinned
`imgproxy-base` Docker image, so you don't need every dependency installed locally.

- **Recommended: Docker + the devcontainer.** You'll need Docker with the Compose
  plugin (`docker compose`) on your host, since `guard_docker` (a `./run` helper)
  uses `docker compose run` against `.devcontainer/docker-compose.yml` to
  re-invoke a task inside the container. See
  [`.devcontainer/oss/README.md`](.devcontainer/oss/README.md) for setup.
- **Without Docker:** install `vips`, `clang-format`, and `lychee` locally (see
  [`.devcontainer/oss/README.md`](.devcontainer/oss/README.md) for the exact packages).
  `./run build`, `./run fmt`, and `./run upgrade-mod` run directly on your host either way.

Either way, install the git hooks once:

```sh
go tool lefthook install
```

## Getting started

Run `./run` with no arguments to list every available task with a one-line description:

```sh
./run
```

Run `./run help <task>` for a task's full usage.

## Common tasks

- `./run build` — build the imgproxy binary (`./imgproxy` by default).
- `./run run [args...]` — run the built binary (inside the base container), sourcing
  `.imgproxyrc` first if present.
- `./run build-and-run [args...]` — build, then run, both inside the base container so
  the binary always matches the environment it runs in.
- `./run test [go-test-args...]` — run the Go test suite via `gotestsum`.
- `./run lint` — run both `lint-go` (golangci-lint) and `lint-clang` (clang-format).
- `./run fmt` — format Go code with `gofmt -s -w`.
- `./run lychee` — check links in `README.md` and `CHANGELOG.md`.

## Git hooks

`lefthook.yml` wires the above tasks into git hooks:

- **pre-commit**: `./run lint-go`, `./run lint-clang`
- **pre-push**: `./run test`, `./run lychee`

## Maintainer-only tasks

A few tasks are for maintainers cutting releases or updating pinned dependencies, rather
than day-to-day contribution:

- `./run bump-version <X.Y.Z>` — bump the version in `version/version.go` and
  `CHANGELOG.md`.
- `./run update-base-image <new-version>` — update the pinned `imgproxy-base` image
  version across the project.
- `./run upgrade-mod`, `./run upgrade-go-tools`, `./run upgrade-gh-actions` — upgrade Go
  dependencies, Go tool directives, and pinned GitHub Actions, respectively.

## Submitting changes

Before opening a pull request, make sure `./run lint` and `./run test` pass (the git
hooks above will catch this for you if installed). Please keep pull requests focused on
a single change so they're easier to review.
