#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Upgrade direct Go dependencies"; }

help() { echo "Usage: ./run upgrade-mod"; }

main() {
  require_tool_go

  go mod tidy
  # shellcheck disable=SC2046 # intentional word-splitting of the module list
  go get $(go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
  go mod tidy
}
