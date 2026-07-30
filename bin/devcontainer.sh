#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Attach a shell to the running devcontainer"; }

help() { echo "Usage: ./run devcontainer"; }

main() {
  run::require_tool devcontainer "devcontainer CLI is required, see https://github.com/devcontainers/cli#npm-install"
  devcontainer exec --workspace-folder "$PROJECT_ROOT" bash
}
