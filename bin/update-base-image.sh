#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Update the pinned base image version across the project"; }

help() {
  cat <<'EOF'
Usage: ./run update-base-image <new-version>

Pulls ghcr.io/imgproxy/imgproxy-base:<new-version> and updates every place
that pins the base image version: GitHub Actions workflows,
.devcontainer/docker-compose.yml, and docker/Dockerfile. guard_docker (in
.runrc) reads the image straight from docker-compose.yml, so it needs no
update of its own. Finally, greps the repo to confirm the old version
string doesn't linger anywhere it shouldn't.

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

  # Same image guard_docker uses (.devcontainer/docker-compose.yml's "image"),
  # minus its tag, so this stays in sync without hardcoding the repo path.
  local image old_version
  image="$(devcontainer_image)"
  old_version="${image##*:}"
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

  local compose_yml
  compose_yml="$(compose_file)"
  if [ -f "$compose_yml" ]; then
    sed -i.bak "s|\(image: .*$image:\).*|\1$new_version|" "$compose_yml"
    rm -f "${compose_yml}.bak"
    run::msg_ok "updated $compose_yml"
  fi

  if [ -f "$dockerfile" ]; then
    sed -i.bak "s|\(ARG BASE_IMAGE_VERSION=\"\).*\"|\1$new_version\"|" "$dockerfile"
    rm -f "${dockerfile}.bak"
    run::msg_ok "updated $dockerfile"
  fi

  if [ "$old_version" != "$new_version" ]; then
    run::msg_info "checking for leftover references to $old_version..."
    local leftover
    leftover="$(git -C "$PROJECT_ROOT" grep -F -l -- "$old_version" || true)"
    if [ -n "$leftover" ]; then
      run::msg_err "old version $old_version still referenced in:"
      printf '%s\n' "$leftover" >&2
      return 1
    fi
    run::msg_ok "no leftover references to $old_version"
  fi

  printf '\n'
  run::msg_ok "base image version updated to $new_version"
}
