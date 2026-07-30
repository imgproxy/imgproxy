#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Build the imgproxy binary"; }

help() {
  cat <<'EOF'
Usage: ./run build [go-build-args...]

Builds the imgproxy binary. If -o is not passed via go-build-args, the
binary defaults to ./imgproxy.

Example:
  ./run build -o /tmp/imgproxy
EOF
}

main() {
  require_tool_go
  go build -v -o "$BINARY" "$@" "$SRCDIR"
}
