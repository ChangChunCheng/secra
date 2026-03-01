# SECRA Monorepo Root Makefile

.PHONY: help backend-build backend-test frontend-build frontend-test docker-up docker-down clean

help: ## Show this help message
	@echo "SECRA Monorepo Commands"
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
