DB_DSN=postgres://postgres:postgrespassword@localhost:5433/bus_booking?sslmode=disable
MIGRATE=./bin/migrate

.PHONY: run build db-up db-down db-reset migrate-up migrate-down migrate-status clean kill-port

## run: start the API server (connects to local Postgres on 5433)
run:
	DB_PORT=5433 go run ./cmd/api

## build: compile the binary into bin/api
build:
	go build -o bin/api ./cmd/api

## db-up: start the Postgres container
db-up:
	docker compose up -d postgres

## db-down: stop all containers (data is preserved)
db-down:
	docker compose down

## db-reset: stop containers AND delete all data (clean slate)
db-reset:
	docker compose down -v

## kill-port: free port 8080 if something is stuck
kill-port:
	@lsof -i :8080 -t | xargs kill -9 2>/dev/null || echo "port 8080 is already free"

## migrate-up: apply all pending migrations
migrate-up:
	$(MIGRATE) -path ./migrations -database "$(DB_DSN)" up

## migrate-down: roll back the last migration
migrate-down:
	$(MIGRATE) -path ./migrations -database "$(DB_DSN)" down 1

## migrate-status: show migration status
migrate-status:
	$(MIGRATE) -path ./migrations -database "$(DB_DSN)" version

## clean: remove compiled binaries
clean:
	rm -rf bin/api

## help: show available commands
help:
	@echo "Available commands:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
