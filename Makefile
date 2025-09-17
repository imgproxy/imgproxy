# imgproxy Makefile

BINARY := ./imgproxy

GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOFMT := gofmt
GOLINT := golangci-lint
GOTESTSUM := gotestsum
SRCDIR := .
RCFILE := ./.imgproxyrc
BREW_PREFIX :=

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
.PHONY: test
test:
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
.PHONY: lint
lint:
	@$(GOLINT) run

# Upgrade direct Go dependencies
.PHONY: upgrade
upgrade:
	@$(GOCMD) mod tidy
	@$(GOCMD) get $$($(GOCMD) list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all)
	@$(GOCMD) mod tidy

# Make any unknown target do nothing to avoid "up to date" messages
.PHONY: FORCE
%: FORCE
	@:
