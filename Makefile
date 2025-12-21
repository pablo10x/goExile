# Exile Project - Developer Experience Makefile

# --- Configuration ---
BINARY_SERVER := server-bin
BINARY_SPAWNER := spawner-bin
GOLINT := $(shell which golangci-lint 2>/dev/null || echo $(shell go env GOPATH)/bin/golangci-lint)
NODE_BIN := npm

# Colors for terminal output
BLUE   := \033[34m
GREEN  := \033[32m
RED    := \033[31m
YELLOW := \033[33m
CYAN   := \033[36m
RESET  := \033[0m

# --- Default Goal ---
.DEFAULT_GOAL := help

.PHONY: help
help: ## Display this help message
	@echo "$(BLUE)Exile Project - available commands:$(RESET)"
	@echo "  $(CYAN)install$(RESET)          - Install all dependencies"
	@echo "  $(CYAN)dev-server$(RESET)       - Run master server in development mode"
	@echo "  $(CYAN)dev-spawner$(RESET)      - Run spawner in development mode"
	@echo "  $(CYAN)dev-frontend$(RESET)     - Run web dashboard in development mode"
	@echo "  $(CYAN)build$(RESET)            - Build all components"
	@echo "  $(CYAN)format$(RESET)           - Format all source code"
	@echo "  $(CYAN)lint$(RESET)             - Run all linters"
	@echo "  $(CYAN)test$(RESET)             - Run all tests"
	@echo "  $(CYAN)clean$(RESET)            - Remove build artifacts"
	@echo "  $(CYAN)check-all$(RESET)        - Run format -> lint -> test -> build"

# --- Installation ---
.PHONY: install
install: install-backend install-frontend ## Install all dependencies

.PHONY: install-backend
install-backend: ## Install Go dependencies
	@echo "$(YELLOW)📦 Installing Backend dependencies...$(RESET)"
	@cd server && go mod tidy
	@cd spawner && go mod tidy

.PHONY: install-frontend
install-frontend: ## Install Node.js dependencies
	@echo "$(YELLOW)📦 Installing Frontend dependencies...$(RESET)"
	@cd web-dashboard && $(NODE_BIN) install

# --- Development ---
.PHONY: dev-server
dev-server: ## Run master server in development mode
	@echo "$(GREEN)🚀 Starting Master Server...$(RESET)"
	@cd server && go run .

.PHONY: dev-spawner
dev-spawner: ## Run spawner in development mode
	@echo "$(GREEN)🚀 Starting Spawner...$(RESET)"
	@cd spawner && go run .

.PHONY: dev-frontend
dev-frontend: ## Run web dashboard in development mode
	@echo "$(GREEN)🚀 Starting Web Dashboard...$(RESET)"
	@cd web-dashboard && $(NODE_BIN) run dev

# --- Build ---
.PHONY: build
build: build-server build-spawner build-frontend ## Build all components

.PHONY: build-server
build-server: ## Build Master Server binary
	@echo "$(YELLOW)🏗️  Building Master Server...$(RESET)"
	@cd server && go build -o ../$(BINARY_SERVER) .

.PHONY: build-spawner
build-spawner: ## Build Spawner binary
	@echo "$(YELLOW)🏗️  Building Spawner...$(RESET)"
	@cd spawner && go build -o ../$(BINARY_SPAWNER) .

.PHONY: build-frontend
build-frontend: ## Build Web Dashboard for production
	@echo "$(YELLOW)🏗️  Building Web Dashboard...$(RESET)"
	@cd web-dashboard && $(NODE_BIN) run build

# --- Linting & Formatting ---
.PHONY: format
format: ## Format all source code
	@echo "$(CYAN)💅 Formatting code...$(RESET)"
	@cd server && go fmt ./...
	@cd spawner && go fmt ./...
	@cd web-dashboard && $(NODE_BIN) run format

.PHONY: lint
lint: lint-backend lint-frontend ## Run all linters

.PHONY: lint-backend
lint-backend: ## Run golangci-lint on Go components
	@echo "$(CYAN)🔍 Linting Server...$(RESET)"
	@cd server && $(GOLINT) run ./...
	@echo "$(CYAN)🔍 Linting Spawner...$(RESET)"
	@cd spawner && $(GOLINT) run ./...

.PHONY: lint-frontend
lint-frontend: ## Run linter on web dashboard
	@echo "$(CYAN)🔍 Linting Frontend (web-dashboard)...$(RESET)"
	@cd web-dashboard && $(NODE_BIN) run lint

.PHONY: check-frontend
check-frontend: ## Run Svelte type-checking
	@echo "$(CYAN)🧪 Type checking Frontend...$(RESET)"
	@cd web-dashboard && $(NODE_BIN) run check

# --- Testing ---
.PHONY: test
test: test-backend test-frontend ## Run all tests

.PHONY: test-backend
test-backend: ## Run Go tests
	@echo "$(GREEN)🧪 Running Backend tests...$(RESET)"
	@cd server && go test -v ./...
	@cd spawner && go test -v ./...

.PHONY: test-frontend
test-frontend: ## Run frontend tests
	@echo "$(GREEN)🧪 Running Frontend tests...$(RESET)"
	@cd web-dashboard && $(NODE_BIN) run test

# --- Cleanup ---
.PHONY: clean
clean: ## Remove build artifacts
	@echo "$(RED)🧹 Cleaning up build artifacts...$(RESET)"
	@rm -f $(BINARY_SERVER) $(BINARY_SPAWNER)
	@rm -rf web-dashboard/.svelte-kit web-dashboard/build

# --- Comprehensive Check ---
.PHONY: check-all
check-all: format lint test build ## Run everything (format -> lint -> test -> build)
	@echo "$(GREEN)✅ All checks passed successfully!$(RESET)"
