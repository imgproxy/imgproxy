#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Build the imgproxy binary, then run it (inside the base container)"; }

help() {
  cat <<'EOF'
Usage: ./run build-and-run [arg1 arg2 ...]

Builds the imgproxy binary and runs it, both inside the base container so
the binary always matches the environment it runs in.
EOF
}

main() {
  guard_docker "$@"
  run::depends_on build
  "$PROJECT_ROOT/run" run "$@"
}
