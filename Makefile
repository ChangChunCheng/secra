# Secra Makefile指挥中心

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

.PHONY: all build clean test docker-up docker-down migrate-up migrate-status version-env

all: build

# 1. 本地构建
build:
	@echo "🛠️  Building Secra binaries locally ($(VERSION))..."
	@mkdir -p bin
	go build -ldflags="$(LDFLAGS)" -o bin/secra-grpc ./cmd/server/grpc.go
	go build -ldflags="$(LDFLAGS)" -o bin/secra-http ./cmd/server/http.go
	go build -ldflags="$(LDFLAGS)" -o bin/secra ./cmd/cli/secra.go

# 2. Docker 构建与启动
docker-up:
	@echo "🐳 Launching Secra in Docker with version $(VERSION)..."
	@# 动态传递构建参数，确保不依赖 .env 文件中的静态定义
	APP_VERSION=$(VERSION) \
	BUILD_DATE=$(BUILD_DATE) \
	GIT_COMMIT=$(GIT_COMMIT) \
	HOSTNAME=$(BUILT_BY) \
	docker compose up -d --build

docker-down:
	docker compose down

clean:
	rm -rf bin/

test:
	go test -v ./...

migrate-up:
	go run cmd/cli/secra.go migrate up

migrate-status:
	go run cmd/cli/secra.go migrate status
