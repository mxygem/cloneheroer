.PHONY: help migrate-up migrate-down migrate-create migrate-version server build test clean

# Default values
MIGRATE_TOOL ?= migrate
MIGRATIONS_DIR ?= migrations
DATABASE_URL ?= postgres://localhost/cloneheroer?sslmode=disable
PORT ?= 3000
WATCH_DIR ?= /tmp/clonehero-screenshots
BUILD_DIR ?= bin
BINARY_NAME ?= cloneheroer

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)Clone Hero Score Tracker - Makefile Targets$(NC)"
	@echo ""
	@echo "$(YELLOW)Migration Targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '(migrate|Migration)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Server Targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -E '(server|build|run)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Other Targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | grep -vE '(migrate|Migration|server|build|run)' | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Environment Variables:$(NC)"
	@echo "  DATABASE_URL  - PostgreSQL connection string (default: $(DATABASE_URL))"
	@echo "  PORT          - HTTP server port (default: $(PORT))"
	@echo "  WATCH_DIR     - Directory to watch for screenshots (default: $(WATCH_DIR))"
	@echo "  MIGRATIONS_DIR - Migrations directory (default: $(MIGRATIONS_DIR))"

# Migration targets
migrate-up: ## Run all pending migrations (up)
	@echo "$(GREEN)Running migrations up...$(NC)"
	@if ! command -v $(MIGRATE_TOOL) > /dev/null; then \
		echo "$(RED)Error: migrate tool not found. Install with:$(NC)"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	echo "$(DATABASE_URL)"
	echo "$(MIGRATIONS_DIR)"
	echo "$(MIGRATE_TOOL)"

	@$(MIGRATE_TOOL) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up
	@echo "$(GREEN)✓ Migrations completed$(NC)"

migrate-down: ## Rollback the last migration (down)
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	@if ! command -v $(MIGRATE_TOOL) > /dev/null; then \
		echo "$(RED)Error: migrate tool not found. Install with:$(NC)"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	@$(MIGRATE_TOOL) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1
	@echo "$(GREEN)✓ Migration rolled back$(NC)"

migrate-down-all: ## Rollback all migrations
	@echo "$(YELLOW)Rolling back all migrations...$(NC)"
	@if ! command -v $(MIGRATE_TOOL) > /dev/null; then \
		echo "$(RED)Error: migrate tool not found. Install with:$(NC)"; \
		echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
	@$(MIGRATE_TOOL) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down -all
	@echo "$(GREEN)✓ All migrations rolled back$(NC)"

migrate-version: ## Show current migration version
	@if ! command -v $(MIGRATE_TOOL) > /dev/null; then \
		echo "$(RED)Error: migrate tool not found$(NC)"; \
		exit 1; \
	fi
	@$(MIGRATE_TOOL) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

migrate-force: ## Force migration to a specific version (use VERSION=n)
	@if [ -z "$(VERSION)" ]; then \
		echo "$(RED)Error: VERSION is required. Usage: make migrate-force VERSION=1$(NC)"; \
		exit 1; \
	fi
	@if ! command -v $(MIGRATE_TOOL) > /dev/null; then \
		echo "$(RED)Error: migrate tool not found$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Forcing migration to version $(VERSION)...$(NC)"
	@$(MIGRATE_TOOL) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(VERSION)

migrate-create: ## Create a new migration (use NAME=description)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required. Usage: make migrate-create NAME=add_indexes$(NC)"; \
		exit 1; \
	fi
	@if ! command -v $(MIGRATE_TOOL) > /dev/null; then \
		echo "$(RED)Error: migrate tool not found$(NC)"; \
		exit 1; \
	fi
	@$(MIGRATE_TOOL) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)
	@echo "$(GREEN)✓ Created migration files in $(MIGRATIONS_DIR)/$(NC)"

# Server targets
server: ## Start the development server
	@echo "$(GREEN)Starting server...$(NC)"
	@echo "$(YELLOW)Configuration:$(NC)"
	@echo "  DATABASE_URL: $(DATABASE_URL)"
	@echo "  PORT: $(PORT)"
	@echo "  WATCH_DIR: $(WATCH_DIR)"
	@echo ""
	@WATCH_DIR=$(WATCH_DIR) \
	 DATABASE_URL=$(DATABASE_URL) \
	 PORT=$(PORT) \
	 go run cmd/server/main.go

server-build: ## Build the server binary
	@echo "$(GREEN)Building server...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go
	@echo "$(GREEN)✓ Binary built: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

server-run: server-build ## Build and run the server binary
	@echo "$(GREEN)Running server...$(NC)"
	@WATCH_DIR=$(WATCH_DIR) \
	 DATABASE_URL=$(DATABASE_URL) \
	 PORT=$(PORT) \
	 ./$(BUILD_DIR)/$(BINARY_NAME)

# Development targets
test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	@go test ./... -v

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test ./... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)Warning: golangci-lint not found, skipping$(NC)"; \
	fi

fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✓ Code formatted$(NC)"

tidy: ## Run go mod tidy
	@echo "$(GREEN)Running go mod tidy...$(NC)"
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies updated$(NC)"

# Utility targets
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Cleaned$(NC)"

deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

check-env: ## Check required environment variables
	@echo "$(GREEN)Checking environment...$(NC)"
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "$(RED)Error: DATABASE_URL not set$(NC)"; \
		exit 1; \
	fi
	@if [ -z "$(WATCH_DIR)" ]; then \
		echo "$(RED)Error: WATCH_DIR not set$(NC)"; \
		exit 1; \
	fi
	@echo "$(GREEN)✓ Environment check passed$(NC)"

install-migrate: ## Install the migrate tool
	@echo "$(GREEN)Installing migrate tool...$(NC)"
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "$(GREEN)✓ Migrate tool installed$(NC)"

generate-test-images: ## Generate test images for parser testing
	@echo "$(GREEN)Generating test images...$(NC)"
	@cd scripts && go run generate_test_images.go
	@echo "$(GREEN)✓ Test images generated$(NC)"

# Default target
.DEFAULT_GOAL := help

