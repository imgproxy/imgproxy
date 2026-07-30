#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Build a distro package via docker/imgproxy-build-package"; }

help() {
  cat <<'EOF'
Usage: ./run build-package <package-type> <target-dir>

Forwards all arguments to docker/imgproxy-build-package. That script
expects to run inside a built imgproxy image (it reads from /opt/imgproxy),
not on the host or in the base dev container.

  package-type: deb, rpm, or tar
  target-dir:   directory to save the package to
EOF
}

main() {
  "$PROJECT_ROOT/docker/imgproxy-build-package" "$@"
}
