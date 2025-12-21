# Exile Project Management

GOLINT := $(shell which golangci-lint 2>/dev/null || echo $(shell go env GOPATH)/bin/golangci-lint)

.PHONY: lint-backend lint-frontend lint check-all format

# Linting for Go components
lint-backend:
	@echo "🔍 Linting Server..."
	@cd server && $(GOLINT) run ./...
	@echo "🔍 Linting Spawner..."
	@cd spawner && $(GOLINT) run ./...

# Linting for Svelte component
lint-frontend:
	@echo "🔍 Linting Frontend (web-dashboard)..."
	@cd web-dashboard && npm run lint

# Svelte check (Type checking)
check-frontend:
	@echo "🔍 Type checking Frontend..."
	@cd web-dashboard && npm run check

# Run all linting and checks
lint: lint-backend lint-frontend check-frontend

# Format all code
format:
	@echo "💅 Formatting code..."
	@cd server && go fmt ./...
	@cd spawner && go fmt ./...
	@cd web-dashboard && npm run format

# Run everything before commit
check-all: format lint
	@echo "✅ All checks passed!"
