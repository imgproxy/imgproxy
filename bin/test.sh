#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Run the Go test suite (inside the base container)"; }

help() {
  cat <<'EOF'
Usage: ./run test [go-test-args...]

Runs the Go test suite via gotestsum. Any extra arguments are forwarded to
`go test` (gotestsum expects them after its own "--" separator, which this
task inserts for you).

Example:
  ./run test -run TestFoo
EOF
}

main() {
  guard_docker "$@"
  require_tool_go
  go tool gotestsum -- ./... "$@"
}
