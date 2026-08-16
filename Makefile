.PHONY: dev build test lint migrate

dev:
	docker compose -f compose.dev.yaml up --build

build:
	go build -o bin/neetrack ./cmd/neetrack

test:
	go test ./...

lint:
	golangci-lint run ./...
	cd web && npm run lint && npm run typecheck

migrate:
	go run ./cmd/neetrack migrate
