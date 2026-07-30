#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Bump the imgproxy version in version.go and CHANGELOG.md"; }

help() {
  cat <<'EOF'
Usage: ./run bump-version <X.Y.Z>

Updates version/version.go and CHANGELOG.md to the given version, then
prints the git commands needed to commit, tag, and push the bump.

Example:
  ./run bump-version 4.0.13
EOF
}

main() {
  local new_version="${1:-}"

  if [ -z "$new_version" ]; then
    echo "error: missing required argument NEW_VERSION (e.g. 4.0.13)" >&2
    help >&2
    return 1
  fi

  echo "$new_version" | grep -qE '^[0-9]+\.[0-9]+\.[0-9]+$' || {
    echo "error: invalid version format '$new_version' (expected X.Y.Z)" >&2
    return 1
  }

  run::msg_info "bumping version to $new_version"

  sed -i.bak "s/const Version = \".*\"/const Version = \"$new_version\"/" "$PROJECT_ROOT/version/version.go"
  rm -f "$PROJECT_ROOT/version/version.go.bak"
  run::msg_ok "updated version/version.go"

  sed -i.bak "s/## \[$new_version] .*/## [$new_version] - $(date +%Y-%m-%d)/" "$PROJECT_ROOT/CHANGELOG.md"
  rm -f "$PROJECT_ROOT/CHANGELOG.md.bak"
  run::msg_ok "updated CHANGELOG.md"

  printf '\nTo complete the version bump:\n'
  printf '  git add .\n'
  printf '  git commit -am "Bump version to %s"\n' "$new_version"
  printf '  git push\n'
  printf '  git tag v%s\n' "$new_version"
  printf '  git push origin v%s\n' "$new_version"
}
