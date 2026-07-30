#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Format Go code"; }

help() { echo "Usage: ./run fmt"; }

main() {
  run::require_tool gofmt "gofmt is required (it ships with Go), see https://go.dev/doc/install"
  gofmt -s -w .
}
