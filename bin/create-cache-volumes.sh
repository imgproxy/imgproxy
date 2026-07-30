#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Create the docker cache volume(s) referenced by devcontainer.json"; }

help() {
  cat <<'EOF'
Usage: ./run create-cache-volumes

Ensures every volume referenced by .devcontainer/devcontainer.json's
"mounts" (source=...,volume-subpath=...) exists, with each subpath
pre-created inside it. Docker requires a volume-subpath to already exist
before it can be mounted there, so this must run before anything mounts
those volumes -- both the devcontainer (via initializeCommand) and
guard_docker (in .runrc) call this task for that reason, rather than
each creating the volume(s) their own way.
EOF
}

main() {
  require_tool_docker

  local -a volume_subpaths=()
  local mount
  while IFS= read -r mount; do
    [ -n "$mount" ] || continue

    local vol="" subpath="" kv
    local -a kv_pairs=()
    IFS=',' read -ra kv_pairs <<< "$mount"
    for kv in "${kv_pairs[@]}"; do
      case "$kv" in
        source=*) vol="${kv#source=}" ;;
        volume-subpath=*) subpath="${kv#volume-subpath=}" ;;
      esac
    done
    [ -n "$vol" ] && [ -n "$subpath" ] && volume_subpaths+=("$vol:$subpath")
  done < <(jq -r '.mounts[] | select(type == "string")' <<< "$(devcontainer_json)")

  if [ "${#volume_subpaths[@]}" -eq 0 ]; then
    run::msg_skip "no volume mounts with a volume-subpath found in devcontainer.json"
    return 0
  fi

  local vol
  for vol in $(printf '%s\n' "${volume_subpaths[@]}" | cut -d: -f1 | sort -u); do
    docker volume create "$vol" >/dev/null

    local -a subpath_names=() subpath_dirs=()
    local entry
    for entry in "${volume_subpaths[@]}"; do
      case "$entry" in
        "$vol":*)
          subpath_names+=("${entry#*:}")
          subpath_dirs+=("/data/${entry#*:}")
          ;;
      esac
    done

    docker run --rm -v "$vol:/data" busybox sh -c "mkdir -p ${subpath_dirs[*]}"
    run::msg_ok "created/updated docker volume $vol, subpaths: (${subpath_names[*]})"
  done
}
