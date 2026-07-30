#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Update the pinned base image version across the project"; }

help() {
  cat <<'EOF'
Usage: ./run update-base-image <new-version>

Pulls ghcr.io/imgproxy/imgproxy-base:<new-version> and updates every place
that pins the base image version: Makefile, GitHub Actions workflows,
.devcontainer/devcontainer.json, and docker/Dockerfile. guard_docker (in
.runrc) reads the image straight from devcontainer.json, so it needs no
update of its own.

Example:
  ./run update-base-image v4.2.0
EOF
}

main() {
  local new_version="${1:-}"

  if [ -z "$new_version" ]; then
    run::msg_err "missing required version argument (e.g. v4.2.0)"
    help >&2
    return 1
  fi

  case "$new_version" in
    v[0-9]*.[0-9]*.[0-9]*) ;;
    *) run::msg_err "version must be in format vX.Y.Z (e.g. v4.2.0)"; return 1 ;;
  esac

  require_tool_docker

  # Same image guard_docker uses (.devcontainer/devcontainer.json's "image"),
  # minus its tag, so this stays in sync without hardcoding the repo path.
  local image
  image="$(devcontainer_image)"
  image="${image%:*}"

  local workflows_dir="$PROJECT_ROOT/.github/workflows"
  local dockerfile="$PROJECT_ROOT/docker/Dockerfile"

  run::msg_info "pulling new base image $image:$new_version..."
  docker pull "$image:$new_version"

  run::msg_info "updating base image version to $new_version..."

  if [ -d "$workflows_dir" ]; then
    local workflow
    for workflow in "$workflows_dir"/*.yml "$workflows_dir"/*.yaml; do
      [ -f "$workflow" ] || continue
      # Match on the variable being replaced, not the image string: the
      # registry and repo are separate env vars (only concatenated at
      # workflow-run time via GH Actions interpolation), so the full image
      # string never appears literally in the file.
      grep -q '^[[:space:]]*CONTAINER_IMAGE_TAG:' "$workflow" || continue
      sed -i.bak "s|\(CONTAINER_IMAGE_TAG: \).*|\1$new_version|g" "$workflow"
      rm -f "${workflow}.bak"
      run::msg_ok "updated $workflow"
    done
  fi

  if [ -L "$PROJECT_ROOT/.devcontainer/devcontainer.json" ]; then
    local devcontainer_file
    devcontainer_file="$(readlink "$PROJECT_ROOT/.devcontainer/devcontainer.json")"
    case "$devcontainer_file" in
      /*) ;;
      *) devcontainer_file="$PROJECT_ROOT/.devcontainer/$devcontainer_file" ;;
    esac
    sed -i.bak "s|\(\"image\": \".*$image:\).*\"|\1$new_version\"|" "$devcontainer_file"
    rm -f "${devcontainer_file}.bak"
    run::msg_ok "updated $devcontainer_file"
  fi

  if [ -f "$dockerfile" ]; then
    sed -i.bak "s|\(ARG BASE_IMAGE_VERSION=\"\).*\"|\1$new_version\"|" "$dockerfile"
    rm -f "${dockerfile}.bak"
    run::msg_ok "updated $dockerfile"
  fi

  printf '\n'
  run::msg_ok "base image version updated to $new_version"
}
