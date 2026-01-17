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
	docker compose up --build

stop:
	docker compose down

# 4. Database Migrations (Optional - if you use 'golang-migrate')
# You can add migration commands here later

# 5. Clean up
clean:
	docker compose down -v
	rm -rf internal/db/*

.PHONY: init gen dev stop clean