include .env
export POSTGRES_USER POSTGRES_PASSWORD POSTGRES_DB POSTGRES_HOST

DATABASE_URL=postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):5432/$(POSTGRES_DB)?sslmode=disable
MIGRATE=migrate -path ./migrations -database $(DATABASE_URL)
APP_NAME=pr-server

.PHONY: build run test lint migrate-up migrate-down migrate-version migrate-force

build:
	go build -o bin/$(APP_NAME) ./cmd/PR/main.go

run: build
	POSTGRES_USER=$(POSTGRES_USER) \
	POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
	POSTGRES_DB=$(POSTGRES_DB) \
	POSTGRES_HOST=$(POSTGRES_HOST) \
	bin/$(APP_NAME)

test:
	go test -race ./...

lint:
	golangci-lint run

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

migrate-version:
	$(MIGRATE) version

migrate-force:
	$(MIGRATE) force $(ver)
