.PHONY: help install-tools sqlc migrate-up migrate-down migrate-status run build clean test docker-up docker-down

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-tools: ## Install required tools (sqlc, goose)
	@echo "Installing sqlc..."
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "Installing goose..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@echo "Tools installed successfully!"

sqlc: ## Generate sqlc code
	@echo "Generating sqlc code..."
	@~/go/bin/sqlc generate
	@echo "Code generated successfully!"

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@go run main.go migrate-up || echo "Run 'make run' to apply migrations automatically"

migrate-down: ## Rollback last migration
	@echo "Rolling back last migration..."
	@~/go/bin/goose -dir database/migrations postgres "${DATABASE_URL}" down

migrate-status: ## Check migration status
	@~/go/bin/goose -dir database/migrations postgres "${DATABASE_URL}" status

run: ## Run the application
	@echo "Starting application..."
	@go run main.go

build: ## Build the application
	@echo "Building application..."
	@go build -o bin/log-server main.go
	@echo "Build complete! Binary: bin/log-server"

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf internal/db/*.go
	@echo "Clean complete!"

test: ## Run tests
	@echo "Running tests..."
	@go test -v ./...

docker-up: ## Start PostgreSQL with Docker Compose
	@echo "Starting PostgreSQL..."
	@docker compose up -d
	@echo "PostgreSQL started!"

docker-down: ## Stop PostgreSQL
	@echo "Stopping PostgreSQL..."
	@docker compose down
	@echo "PostgreSQL stopped!"

docker-logs: ## View PostgreSQL logs
	@docker compose logs -f postgres

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated!"

dev: docker-up ## Start development environment (PostgreSQL + App)
	@echo "Waiting for PostgreSQL to be ready..."
	@sleep 3
	@make run

all: deps install-tools sqlc build ## Install tools, generate code, and build

.DEFAULT_GOAL := help