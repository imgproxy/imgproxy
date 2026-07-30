#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Regenerate docs/sitemap.txt from docs/_sidebar.md"; }

help() { echo "Usage: ./run docs-sitemap"; }

main() {
  local re='^\* \[.+\]\((.+)\)'
  echo "https://docs.imgproxy.net" > "$PROJECT_ROOT/docs/sitemap.txt"
  grep -E "$re" "$PROJECT_ROOT/docs/_sidebar.md" \
    | sed -E "s|$re|https://docs.imgproxy.net/\1|" >> "$PROJECT_ROOT/docs/sitemap.txt"
}
