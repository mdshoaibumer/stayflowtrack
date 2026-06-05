.PHONY: help build run test lint migrate-up migrate-down sqlc docker-up docker-down

# Default target
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ==================== Development ====================

run: ## Run the server locally
	go run ./cmd/server

build: ## Build the server binary
	go build -o bin/server ./cmd/server

test: ## Run all tests
	go test ./... -v -race -count=1

test-coverage: ## Run tests with coverage
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	gofmt -w .
	goimports -w .

# ==================== Database ====================

DB_URL ?= postgres://stayflow:stayflow_dev_password@localhost:5432/stayflow_track?sslmode=disable

migrate-up: ## Run database migrations
	migrate -path migrations -database "$(DB_URL)" up

migrate-down: ## Rollback last migration
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-create: ## Create a new migration (usage: make migrate-create name=migration_name)
	migrate create -ext sql -dir migrations -seq $(name)

migrate-force: ## Force migration version (usage: make migrate-force version=1)
	migrate -path migrations -database "$(DB_URL)" force $(version)

# ==================== Code Generation ====================

sqlc: ## Generate sqlc code
	cd sqlc && sqlc generate

# ==================== Docker ====================

docker-up: ## Start all services with Docker Compose
	docker compose up -d

docker-down: ## Stop all services
	docker compose down

docker-build: ## Build Docker images
	docker compose build

docker-logs: ## View logs
	docker compose logs -f

docker-migrate: ## Run migrations in Docker
	docker compose run --rm migrate

# ==================== Utilities ====================

clean: ## Clean build artifacts
	rm -rf bin/ coverage.out coverage.html

deps: ## Download dependencies
	go mod download
	go mod tidy

tools: ## Install development tools
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
