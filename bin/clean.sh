#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Remove build artifacts"; }

help() { echo "Usage: ./run clean"; }

main() {
  require_tool_go
  echo "$PKG_CONFIG_PATH"
  go clean
  rm -f "$BINARY"
}
