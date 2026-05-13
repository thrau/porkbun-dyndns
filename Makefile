all: build

build:
	go build -o bin/porkbun-dyndns ./cmd/porkbun-dyndns

clean:
	rm -rf bin

test:
	go test -v ./...

.PHONY: all build clean test
