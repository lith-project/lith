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
	@printf 'corpus: not implemented yet — corpus integrity check tracked in #36\n' >&2
	@exit 1

check:
	$(MAKE) build
	$(MAKE) test
	$(MAKE) lint
	$(MAKE) conformance
	$(MAKE) corpus

clean:
	rm -f lithd lith
