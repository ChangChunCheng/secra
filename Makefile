# Secra Makefile 指挥中心

APP_NAME := secra
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILT_BY := $(shell hostname)
GO_OS := $(shell go env GOOS)
GO_ARCH := $(shell go env GOARCH)

# 注入路径
PKG := gitlab.com/jacky850509/secra/internal/config

LDFLAGS := -X $(PKG).Version=$(VERSION) \
           -X $(PKG).BuildDate=$(BUILD_DATE) \
           -X $(PKG).Commit=$(GIT_COMMIT) \
           -X $(PKG).BuiltBy=$(BUILT_BY) \
           -X $(PKG).OS=$(GO_OS) \
           -X $(PKG).Arch=$(GO_ARCH) \
           -s -w

.PHONY: all build clean test docker-up docker-down migrate-up migrate-status backup restore help

all: build

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# 1. 本地构建 (Consolidated)
build: ## Build Secra binaries locally
	@echo "🛠️  Building Secra binaries locally ($(VERSION))..."
	@mkdir -p bin
	go build -ldflags="$(LDFLAGS)" -o bin/secra-server ./cmd/server/main.go
	go build -ldflags="$(LDFLAGS)" -o bin/secra ./cmd/cli/secra.go

# 2. Docker 运行
docker-up: ## Build and start the consolidated server in Docker
	@echo "🐳 Launching Secra Monolith in Docker with version $(VERSION)..."
	APP_VERSION=$(VERSION) \
	BUILD_DATE=$(BUILD_DATE) \
	GIT_COMMIT=$(GIT_COMMIT) \
	HOSTNAME=$(BUILT_BY) \
	docker compose up -d --build

docker-down: ## Stop and remove all containers
	docker compose down

# 3. 数据库维护
migrate-up: ## Run database migrations
	docker compose exec server secra migrate up

migrate-status: ## Check migration status
	docker compose exec server secra migrate status

# 4. 备份与还原
backup: ## Create a full system backup (Usage: make backup OUT=./backups)
	@chmod +x ./scripts/backup.sh
	./scripts/backup.sh $(or $(OUT),./backups)

restore: ## Restore system from a backup file (Usage: make restore FILE=./backups/xxx.tar.gz)
	@chmod +x ./scripts/restore.sh
	./scripts/restore.sh $(FILE)

clean: ## Remove build artifacts
	rm -rf bin/

test: ## Run all tests
	go test -v ./...
