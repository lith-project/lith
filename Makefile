.PHONY: build test lint fmt conformance corpus check clean

build:
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w .
	goimports -w .

conformance:
	go run ./tools/conformance/boundary

corpus:
	go run ./tools/conformance/corpus

check:
	$(MAKE) build
	$(MAKE) test
	$(MAKE) lint
	$(MAKE) conformance

clean:
	rm -f lithd lith
