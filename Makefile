.PHONY: run test check build fmt

run:
	go run ./cmd/bandia

test:
	go test -race ./...

check:
	go vet ./...
	go test -race ./...

build:
	mkdir -p bin
	go build -o bin/bandia ./cmd/bandia

fmt:
	gofmt -w cmd internal
