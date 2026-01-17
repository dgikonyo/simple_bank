# --- Variables ---
DB_URL=postgresql://root_user:root_secret@localhost:5432/bank_db?sslmode=disable
BINARY_NAME=main

# --- Commands ---

# 1. Initialize the project (Run this first)
init:
	go mod init simple_bank || true
	go mod tidy

# 2. Database Code Generation
gen:
	sqlc generate

# 3. Docker Commands
dev:
	docker compose up --build --remove-orphans

stop:
	docker compose down --remove-orphans

# 5. Clean up
clean:
	docker compose down -v
	rm -rf internal/db/*
	docker ps -q --filter "ancestor=cosmtrek/air" | xargs -r docker rm -f

# A "Nuke" command for when things get really messy
reset:
	docker compose down -v --remove-orphans
	docker system prune -f

# Start only the test database
test-db-up:
	docker compose up -d postgres_test

# Run tests
test:
	go test -v -cover ./internal/db/...

# Stop the test database
test-db-down:
	docker compose stop postgres_test

.PHONY: init gen dev stop clean test test-db-up test-db-down clean reset