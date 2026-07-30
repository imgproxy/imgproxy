#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Scan for known vulnerabilities (inside the base container)"; }

help() { echo "Usage: ./run govulncheck"; }

main() {
  guard_docker "$@"
  require_tool_go
  go tool govulncheck -scan symbol ./...
}
