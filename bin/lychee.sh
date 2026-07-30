#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Check links in README.md and CHANGELOG.md (inside the base container)"; }

help() { echo "Usage: ./run lychee"; }

main() {
  guard_docker "$@"
  run::require_tool lychee "lychee is required, see https://github.com/lycheeverse/lychee#installation"
  lychee README.md CHANGELOG.md
}
