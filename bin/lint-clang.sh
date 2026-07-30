#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Lint C code (inside the base container)"; }

help() { echo "Usage: ./run lint-clang"; }

main() {
  guard_docker "$@"
  run::require_tool clang-format "clang-format is required, install via your package manager (e.g. 'brew install clang-format' or 'apt install clang-format') or see https://releases.llvm.org/download.html"
  find . -not -path "./.tmp/*" -not -path "./.git/*" \( -iname "*.h" -o -iname "*.c" -o -iname "*.cpp" \) \
    | xargs clang-format --dry-run --Werror
}
