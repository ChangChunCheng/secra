# SECRA Monorepo Root Makefile

# Build version information
APP_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILT_BY := $(shell echo "$${USER}@$$(hostname)")
HOSTNAME := $(shell hostname)

export APP_VERSION
export BUILD_DATE
export GIT_COMMIT
export BUILT_BY
export HOSTNAME

.PHONY: help version build build-server build-frontend backend-build backend-test frontend-build frontend-test docker-up docker-down clean

help: ## Show this help message
	@echo "SECRA Monorepo Commands"
	@echo ""
	@echo "Version:"
	@echo "  make version             - Display build version information"
	@echo ""
	@echo "Build:"
	@echo "  make build               - Build all Docker images with version tags"
	@echo "  make build-server        - Build server image only"
	@echo "  make build-frontend      - Build frontend image only"
	@echo ""
	@echo "Backend:"
	@echo "  make backend-build       - Build backend binaries"
	@echo "  make backend-test        - Run backend tests"
	@echo ""
	@echo "Frontend:"
	@echo "  make frontend-build      - Build frontend"
	@echo "  make frontend-test       - Run frontend tests"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up           - Start all services"
	@echo "  make docker-down         - Stop all services"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean               - Clean all build artifacts"

version: ## Display build version information
	@echo "======================================"
	@echo "Secra Build Information"
	@echo "======================================"
	@echo "Version:    $(APP_VERSION)"
	@echo "Build Date: $(BUILD_DATE)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Built By:   $(BUILT_BY)"
	@echo "======================================"

# Backend commands
backend-build: ## Build backend binaries
	@cd backend && $(MAKE) build

backend-test: ## Run backend tests
	@cd backend && $(MAKE) test

# Frontend commands
frontend-build: ## Build frontend
	@cd frontend && npm run build

frontend-test: ## Run frontend tests
	@cd frontend && npm test

# Docker commands
build: version ## Build all Docker images with version information
	@echo ""
	@echo "Building all services..."
	docker compose build

build-server: version ## Build server image only
	@echo ""
	@echo "Building server..."
	docker compose build server

build-frontend: version ## Build frontend image only
	@echo ""
	@echo "Building frontend..."
	docker compose build frontend

docker-up: ## Start all services with Docker Compose
	@cd backend && $(MAKE) docker-up

docker-down: ## Stop all Docker services
	@cd backend && $(MAKE) docker-down

migrate-up: ## Run database migrations
	@cd backend && $(MAKE) migrate-up

migrate-status: ## Check migration status
	@cd backend && $(MAKE) migrate-status

# Clean
clean: ## Clean all build artifacts
	@cd backend && $(MAKE) clean
	@cd frontend && rm -rf .next out
	@echo "✅ All build artifacts cleaned"
