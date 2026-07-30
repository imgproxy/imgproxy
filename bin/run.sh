#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Run the imgproxy binary"; }

help() {
  cat <<'EOF'
Usage: ./run run [arg1 arg2 ...]

Runs the imgproxy binary (inside the base container) forwarding any
arguments. If .imgproxyrc exists at the project root, it is sourced first.

Note: this runs the container-native binary, so it must have been built
inside the container too -- use ./run build-and-run, not a host build
followed by ./run run, or the binary won't execute (wrong OS/arch).
EOF
}

main() {
  guard_docker "$@"

  if [ -f "$IMGPROXY_RCFILE" ]; then
    # shellcheck disable=SC1090
    . "$IMGPROXY_RCFILE"
  fi
  "$BINARY" "$@"
}
