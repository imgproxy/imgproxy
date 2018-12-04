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

test: 
	CGO_LDFLAGS_ALLOW="-s|-w" go test -v

# dev run
dev: build
	./imgproxy
