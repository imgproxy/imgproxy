#!/usr/bin/env bash
# Sourced by ./run — do not execute directly.

description() { echo "Run the SVG test suite against svgo-test-suite (inside the base container)"; }

help() {
  cat <<'EOF'
Usage: ./run test-svg

Downloads the svgo-test-suite archive into .test-data (skipped if it's
already there) and runs TestSvgSuite against it via `go test -v`, bypassing
gotestsum so the suite's per-file logging stays readable.

Source: https://svg.github.io/svgo-test-suite/svgo-test-suite.tar.gz
EOF
}

SVGO_TEST_SUITE_URL="https://svg.github.io/svgo-test-suite/svgo-test-suite.tar.gz"

main() {
  guard_docker "$@"
  require_tool_go
  run::require_tool curl "curl is required to download the svgo-test-suite archive"
  run::require_tool tar "tar is required to extract the svgo-test-suite archive"

  local data_dir="$PROJECT_ROOT/.test-data"
  local suite_dir="$data_dir/svgo-test-suite"

  if [ ! -d "$suite_dir" ]; then
    mkdir -p "$data_dir"

    # Extract into a temp dir first and mv into place on success, so a
    # failed/interrupted download can never leave behind a $suite_dir that
    # looks cached (and gets silently skipped) but is actually incomplete.
    local tmp_dir
    tmp_dir="$(mktemp -d "$data_dir/.svgo-test-suite.XXXXXX")"

    run::msg_ok "downloading svgo-test-suite..."
    curl -fsSL "$SVGO_TEST_SUITE_URL" | tar -xz -C "$tmp_dir"
    mv "$tmp_dir" "$suite_dir"
  fi

  TEST_SVG_PATH="$suite_dir" go test -v ./integration_test/... -run TestSvgSuite "$@"
}
