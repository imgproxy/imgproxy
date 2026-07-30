#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Upgrade go tool dependencies (go.mod tool directives) to latest"; }

help() { echo "Usage: ./run upgrade-go-tools"; }

main() {
  require_tool_go

  local tools
  tools="$(awk '/^tool \(/{flag=1; next} /^\)/{flag=0} flag {print $1}' "$PROJECT_ROOT/go.mod")"

  if [ -z "$tools" ]; then
    run::msg_skip "no tool directives found in go.mod"
    return 0
  fi

  local tool
  while IFS= read -r tool; do
    [ -n "$tool" ] || continue
    run::msg_info "upgrading $tool"
    go get -tool "$tool@latest"
  done <<< "$tools"

  go mod tidy
}
