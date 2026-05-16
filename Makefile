all: build

build:
	go build -o bin/porkbun-dns ./cmd/porkbun-dns
	go build -o bin/porkbun-ddnsd ./cmd/porkbun-ddnsd

clean:
	rm -rf bin/
	rm -rf dist/

dist:
	goreleaser release --snapshot

test:
	go test -v ./...

.PHONY: all build clean dist test
