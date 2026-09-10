.PHONY: run migrate test integration check build fmt init up down

init:
	python3 scripts/init-local.py

up:
	docker compose up --build -d

down:
	docker compose down

migrate:
	go run ./cmd/bandia migrate

integration:
	@test -n "$$BANDIA_TEST_DATABASE_URL" || (echo "Set BANDIA_TEST_DATABASE_URL to a disposable PostgreSQL database"; exit 1)
	go test -race -count=1 ./internal/store

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
