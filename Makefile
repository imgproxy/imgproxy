# imgproxy Makefile

BINARY := ./imgproxy

MAKEFILE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOFMT := gofmt
GOLINT := golangci-lint
CLANG_FORMAT := clang-format
GOTESTSUM := gotestsum
SRCDIR := ./cli
RCFILE := ./.imgproxyrc
BREW_PREFIX :=
DEVROOT_TMP_DIR ?= $(MAKEFILE_DIR).tmp/_dev-root
BASE_IMAGE ?= ghcr.io/imgproxy/imgproxy-base:v4-dev

# Common environment setup for CGO builds
ifneq ($(shell which brew),)
	BREW_PREFIX := $(shell brew --prefix)
endif

# Export CGO environment variables
export CGO_LDFLAGS_ALLOW := -s|-w

# Library paths for Homebrew-installed libraries on macOS
ifdef BREW_PREFIX
	export PKG_CONFIG_PATH := $(PKG_CONFIG_PATH):$(shell brew --prefix libffi)/lib/pkgconfig
	export PKG_CONFIG_PATH := $(PKG_CONFIG_PATH):$(shell brew --prefix libarchive)/lib/pkgconfig
	export PKG_CONFIG_PATH := $(PKG_CONFIG_PATH):$(shell brew --prefix cfitsio)/lib/pkgconfig

	export CGO_LDFLAGS := $(CGO_LDFLAGS) -Wl,-no_warn_duplicate_libraries
endif

# Get build arguments
ifeq (build,$(firstword $(MAKECMDGOALS)))
	BUILD_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
endif

# Get run arguments
ifeq (run,$(firstword $(MAKECMDGOALS)))
	RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
endif
ifeq (build-and-run,$(firstword $(MAKECMDGOALS)))
	RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
endif

# Wrapper action to run commands in Docker
.PHONY: _run-in-docker
_run-in-docker:
ifdef IMGPROXY_IN_BASE_CONTAINER
	@$(MAKE) $(DOCKERCMD)
else
	@mkdir -p ${DEVROOT_TMP_DIR}/.cache ${DEVROOT_TMP_DIR}/go/pkg/mod
	@docker run --init --rm -it \
		-v "$(MAKEFILE_DIR):/workspaces/imgproxy" \
		-v "${DEVROOT_TMP_DIR}/.cache:/root/.cache" \
		-v "${DEVROOT_TMP_DIR}/go/pkg/mod:/root/go/pkg/mod" \
		-w /workspaces/imgproxy \
		-e IMGPROXY_IN_BASE_CONTAINER=1 \
		$(BASE_IMAGE) \
		bash -c "make $(DOCKERCMD)"
endif

# Default target
.PHONY: all
all: build

# Build the binary. If -o is not provided, it defaults to $(BINARY).
#
# Usage:
#	make build -- -o output_name
.PHONY: build
build:
	@$(GOBUILD) -v -o $(BINARY) $(BUILD_ARGS) $(SRCDIR)

# Clean
.PHONY: clean
clean:
	echo $$PKG_CONFIG_PATH
	@$(GOCLEAN)
	rm -f $(BINARY)

# Run imgproxy binary
#
# Usage:
#	make run -- arg1 arg2
#
# If .imgproxyrc exists, it will be sourced before running the binary.
.PHONY: run
run: SHELL := bash
run:
ifneq (,$(wildcard $(RCFILE)))
	@source $(RCFILE) && $(BINARY) $(RUN_ARGS)
else
	@$(BINARY) $(RUN_ARGS)
endif

.PHONY: build-and-run
build-and-run: build run

# Run tests
#
# Usage:
#	make test -- -run FooTest
.PHONY: test _test
test: DOCKERCMD := _test
test: _run-in-docker
_test:
ifneq ($(shell which $(GOTESTSUM)),)
	@$(GOTESTSUM) ./...
else
	@$(GOTEST) -v ./...
endif

# Format code
.PHONY: fmt
fmt:
	@$(GOFMT) -s -w .

# Lint code (requires golangci-lint installed)
.PHONY: lint-go _lint-go
lint-go: DOCKERCMD := _lint-go
lint-go: _run-in-docker
_lint-go:
	@$(GOLINT) run

# Lint C code (requires clang-format installed)
.PHONE: lint-clang _lint-clang
ling-clang: DOCKERCMD := _lint-clang
ling-clang: _run-in-docker
_lint-clang:
	 @find . -not -path "./.tmp/*" -not -path "./.git/*" \( -iname "*.h" -o -iname "*.c" -o -iname "*.cpp" \) | xargs $(CLANG_FORMAT) --dry-run --Werror

# Run all linters
.PHONY: lint
lint: lint-go ling-clang

# Upgrade direct Go dependencies
.PHONY: upgrade
upgrade:
	@$(GOCMD) mod tidy
	@$(GOCMD) get $$($(GOCMD) list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	@$(GOCMD) mod tidy

# Run lychee
.PHONY: lychee _lychee
lychee: DOCKERCMD := _lychee
lychee: _run-in-docker
_lychee:
	lychee docs README.md CHANGELOG.md \
		--exclude localhost \
		--exclude twitter.com \
		--exclude x.com \
		--exclude-path docs/index.html \
		--max-concurrency 50

.PHONY: devcontainer
devcontainer:
	devcontainer exec --workspace-folder $(MAKEFILE_DIR) bash


# Make any unknown target do nothing to avoid "up to date" messages
.PHONY: FORCE
%: FORCE
	@:
