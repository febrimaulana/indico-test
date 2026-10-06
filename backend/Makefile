GOBIN ?= $(shell go env GOPATH)/bin
POSTGRES_PORT ?= 5432
TEST_DATABASE_URL ?= postgres://indico:indico@localhost:$(POSTGRES_PORT)/indico_test?sslmode=disable
COMPOSE = docker compose -f ../docker-compose.yaml

.PHONY: run build lint test test-race test-integration coverage coverage-html swag ensure-test-db

run:
	go run ./cmd/api

build:
	go build ./...

lint:
	go vet ./...

## Unit tests only (integration tests auto-skip without TEST_DATABASE_URL).
test:
	go test -p 1 ./...

test-race:
	go test -p 1 -race -v ./...

## Integration tests need the docker compose postgres service.
test-integration: ensure-test-db
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -p 1 -race -v ./internal/... -run Integration

## 100% coverage gate over ./internal/... (unit + integration, race mode).
coverage: ensure-test-db
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -p 1 -race -covermode=atomic -coverprofile=coverage.out -coverpkg=./internal/... ./internal/...
	@go tool cover -func=coverage.out | tail -1
	@go tool cover -func=coverage.out | tail -1 | awk '{ if ($$3 != "100.0%") { print "COVERAGE GATE GAGAL: total " $$3 " (wajib 100.0%)"; exit 1 } }'

coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html

swag:
	$(GOBIN)/swag init -g cmd/api/main.go -o docs --parseInternal --parseDependency

## Starts postgres and ensures the indico_test database exists.
ensure-test-db:
	POSTGRES_PORT=$(POSTGRES_PORT) $(COMPOSE) up -d postgres
	@until $(COMPOSE) exec -T postgres pg_isready -U indico -d indico >/dev/null 2>&1; do sleep 0.5; done
	@$(COMPOSE) exec -T postgres psql -U indico -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='indico_test'" | grep -q 1 \
		|| $(COMPOSE) exec -T postgres createdb -U indico indico_test
