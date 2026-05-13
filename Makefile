all: build

build:
	go build -o bin/porkbun-dyndns ./cmd/porkbun-dyndns
	go build -o bin/porkbun-dyndnsd ./cmd/porkbun-dyndnsd

clean:
	rm -rf bin

test:
	go test -v ./...

.PHONY: all build clean test
