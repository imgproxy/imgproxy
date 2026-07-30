#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Run all linters (Go and C)"; }

help() { echo "Usage: ./run lint"; }

main() {
  run::depends_on lint-go lint-clang
}
