current_dir := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
vendor      := $(current_dir)/_vendor
goenv       := GOPATH="$(vendor):$(GOPATH)"

all: clean vendorize build

clean:
	rm -rf bin/

vendorize:
	cd $(current_dir)
	GOPATH=$(vendor) go get -d
	find $(vendor) -name ".git" -type d | xargs rm -rf

clean-vendor:
	rm -rf $(vendor)

hard-vendorize: clean-vendor vendorize

build:
	cd $(current_dir)
	$(goenv) go build -v -ldflags '-w -s' -o bin/server
