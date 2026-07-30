#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Pin GitHub Actions to commit SHAs with pinact"; }

help() { echo "Usage: ./run upgrade-gh-actions"; }

main() {
  run::require_tool pinact "pinact is not installed. See installation instructions at https://github.com/suzuki-shunsuke/pinact/blob/main/INSTALL.md"

  pinact run -u --min-age 7
}
