all: build

build:
	go build -o bin/porkbun-dns ./cmd/porkbun-dns
	go build -o bin/porkbun-ddnsd ./cmd/porkbun-ddnsd

clean:
	rm -rf bin/

test:
	go test -v ./...

.PHONY: all build clean test
