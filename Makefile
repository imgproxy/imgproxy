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

# Default target
.PHONY: all
all: build

# Build the binary. If -o is not provided, it defaults to $(BINARY).
#
# Usage:
#	make build -- -o output_name
.PHONY: build
build:
	@args="$(filter-out $@,$(MAKECMDGOALS))"; \
	if echo "$$args" | grep -q "\-o"; then \
		$(GOBUILD) $$args $(SRCDIR); \
	else \
		$(GOBUILD) -o $(BINARY) $$args $(SRCDIR); \
	fi

# Clean
.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f $(BINARY)

# Run tests
#
# Usage:
#	make test -- -run FooTest
.PHONY: test
test:
	@$(GOTEST) ./... $(filter-out $@,$(MAKECMDGOALS))

# Run gotestsum
#
# Usage:
#	make testsum -- -run FooTest
testsum:
	@$(GOTESTSUM) -- $(filter-out $@,$(MAKECMDGOALS))

# Format code
.PHONY: fmt
fmt:
	$(GOFMT) -s -w .

# Lint code (requires golangci-lint installed)
.PHONY: lint
lint:
	$(GOLINT) run ./...

# Make any unknown target do nothing to avoid "up to date" messages
.PHONY: FORCE
%: FORCE
	@:
