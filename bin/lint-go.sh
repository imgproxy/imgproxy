#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Lint Go code (inside the base container)"; }

help() { echo "Usage: ./run lint-go [golangci-lint-args...]"; }

main() {
  guard_docker "$@"
  require_tool_go
  go tool golangci-lint run "$@"
}
