SHELL := /bin/bash

.PHONY: dev up down test tidy fmt gqlgen

# boot the full stack
dev: up

up:
	docker compose up --build

down:
	docker compose down -v

# run gofmt on every go file
fmt:
	find . -name '*.go' -not -path './vendor/*' -print0 | xargs -0 gofmt -w

# run go test with coverage (target 80%)
test:
	go test ./...

# sync modules
tidy:
	go mod tidy

# regenerate GraphQL code from schema
gqlgen:
	cd services/api-gateway && go run github.com/99designs/gqlgen generate
