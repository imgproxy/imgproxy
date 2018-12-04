# Imgproxy makefile


# Default: build
all: dep build

# Install dependencies
dep:
	dep ensure

# Build
build:
	CGO_LDFLAGS_ALLOW="-s|-w" go build -v

# Debug
debug:
	CGO_LDFLAGS_ALLOW="-s|-w" dlv debug --listen=127.0.0.1:2345 --log


# dev run
dev:
	CGO_LDFLAGS_ALLOW="-s|-w" go run -v main.go

run: build
	./imgproxy