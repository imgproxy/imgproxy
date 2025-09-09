# imgproxy Makefile

BINARY := imgproxy

GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOFMT := gofmt
GOLINT := golangci-lint
GOTESTSUM := gotestsum
SRCDIR := ./cli
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

# Default target
.PHONY: all
all: build

# Build the binary. If -o is not provided, it defaults to $(BINARY).
#
# Usage:
#	make build -- -o output_name
.PHONY: build
build:
	@$(GOBUILD) -o $(BINARY) $(filter-out $@,$(MAKECMDGOALS)) $(SRCDIR); \

# Clean
.PHONY: clean
clean:
	echo $$PKG_CONFIG_PATH
	@$(GOCLEAN)
	rm -f $(BINARY)

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

# Make any unknown target do nothing to avoid "up to date" messages
.PHONY: FORCE
%: FORCE
	@:
