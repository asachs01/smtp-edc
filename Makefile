.PHONY: help build build-frontend build-backend dev test test-frontend test-backend lint lint-frontend lint-backend format format-frontend format-backend clean install deps-frontend deps-backend vet staticcheck golangci-lint wails-build

# Default target
help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Installation and setup
install: deps-frontend deps-backend ## Install all dependencies

deps-frontend: ## Install frontend dependencies
	cd frontend && npm install

deps-backend: ## Install backend dependencies
	go mod download

# Build targets
build: build-frontend build-backend build-mcp ## Build all components

build-frontend: ## Build frontend only
	cd frontend && npm run build

build-backend: ## Build backend only
	go build -tags cli -o bin/smtp-edc-cli ./cmd/smtp-edc

build-mcp: ## Build MCP server
	go build -o bin/smtp-edc-mcp ./cmd/mcp-server

wails-build: build-frontend ## Build the Wails desktop application
	wails build

wails-dev: ## Run Wails in development mode
	wails dev

# Development targets
dev: ## Start development mode (Wails dev server)
	wails dev

dev-frontend: ## Start frontend development server
	cd frontend && npm run dev

# Testing targets
test: test-frontend test-backend ## Run all tests

test-frontend: ## Run frontend tests
	cd frontend && npm run test

test-backend: ## Run backend tests
	go test -tags cli -v ./...

test-mcp: ## Run MCP tests
	go test -v ./internal/mcp/...

test-coverage: ## Run backend tests with coverage
	go test -tags cli -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Linting and formatting targets
lint: lint-frontend lint-backend ## Run all linters

lint-frontend: ## Run frontend linting
	cd frontend && npm run lint

lint-backend: vet staticcheck golangci-lint ## Run all backend linting tools

format: format-frontend format-backend ## Format all code

format-frontend: ## Format frontend code
	cd frontend && npm run format

format-backend: ## Format backend code
	go fmt ./...
	goimports -w .

# Backend specific linting tools
vet: ## Run go vet
	go vet -tags cli ./...

staticcheck: ## Run staticcheck
	staticcheck -tags cli ./...

golangci-lint: ## Run golangci-lint
	golangci-lint run --build-tags cli ./...

# Type checking
type-check: ## Run TypeScript type checking
	cd frontend && npm run type-check

# Cleaning targets
clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf build/
	rm -rf frontend/dist/
	rm -f coverage.out coverage.html

clean-deps: ## Clean dependencies
	rm -rf frontend/node_modules/
	go clean -modcache

# Utility targets
check: lint test ## Run linting and tests

validate: ## Validate project setup
	@echo "Checking Go installation..."
	@go version
	@echo "Checking Node.js installation..."
	@node --version
	@echo "Checking npm installation..."
	@npm --version
	@echo "Checking Wails installation..."
	@wails version
	@echo "All tools are installed correctly!"

watch-backend: ## Watch backend files for changes and rebuild
	find . -name '*.go' | entr -r make build-backend

# Git hooks setup
setup-hooks: ## Setup git pre-commit hooks
	@echo "#!/bin/sh" > .git/hooks/pre-commit
	@echo "make lint" >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hooks installed"

# Release targets
release-prep: clean lint test build ## Prepare for release
	@echo "Release preparation complete"

# Docker targets (if needed later)
docker-build: ## Build Docker image
	docker build -t smtp-edc .

# Generate documentation
docs: ## Generate documentation
	go doc -all ./... > docs/api.md
