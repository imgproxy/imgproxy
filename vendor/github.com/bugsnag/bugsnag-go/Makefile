TEST?=./...

default: alldeps test

deps:
	go get -v -d ./...

alldeps:
	go get -v -d -t ./...

updatedeps:
	go get -v -d -u ./...

test: alldeps
	#TODO: 2018-09-20 Not testing the 'errors' package as it relies on some very runtime-specific implementation details.
	# The testing of 'errors' needs to be revisited
	go test . ./gin ./martini ./negroni ./sessions ./headers
	@go vet 2>/dev/null ; if [ $$? -eq 3 ]; then \
		go get golang.org/x/tools/cmd/vet; \
	fi
	@go vet $(TEST) ; if [ $$? -eq 1 ]; then \
		echo "go-vet: Issues running go vet ./..."; \
		exit 1; \
	fi

maze:
	bundle install
	bundle exec bugsnag-maze-runner

ci: alldeps test

bench:
	go test --bench=.*

testsetup:
	gem update --system
	gem install bundler
	bundle install

testplain: testsetup
	bundle exec bugsnag-maze-runner -c features/plain_features

testnethttp: testsetup
	bundle exec bugsnag-maze-runner -c features/net_http_features

testgin: testsetup
	bundle exec bugsnag-maze-runner -c features/gin_features

testmartini: testsetup
	bundle exec bugsnag-maze-runner -c features/martini_features

testnegroni: testsetup
	bundle exec bugsnag-maze-runner -c features/negroni_features

testrevel: testsetup
	bundle exec bugsnag-maze-runner -c features/revel_features

.PHONY: bin checkversion ci default deps generate releasebin test testacc testrace updatedeps testsetup testplain testnethttp testgin testmartini testrevel
