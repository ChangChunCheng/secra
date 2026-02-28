# Secra Makefile

APP_NAME := secra
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILT_BY := $(shell hostname)
GO_OS := $(shell go env GOOS)
GO_ARCH := $(shell go env GOARCH)

# Package path for variable injection
PKG := gitlab.com/jacky850509/secra/internal/config

LDFLAGS := -X $(PKG).Version=$(VERSION) \
           -X $(PKG).BuildDate=$(BUILD_DATE) \
           -X $(PKG).Commit=$(GIT_COMMIT) \
           -X $(PKG).BuiltBy=$(BUILT_BY) \
           -X $(PKG).OS=$(GO_OS) \
           -X $(PKG).Arch=$(GO_ARCH) \
           -s -w

.PHONY: all build clean test docker-up docker-down migrate-up migrate-status

all: build

build:
	@echo "Building Secra binaries..."
	go build -ldflags="$(LDFLAGS)" -o bin/secra-grpc ./cmd/server/grpc.go
	go build -ldflags="$(LDFLAGS)" -o bin/secra-http ./cmd/server/http.go
	go build -ldflags="$(LDFLAGS)" -o bin/secra ./cmd/cli/secra.go

clean:
	rm -rf bin/

test:
	go test -v ./...

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

migrate-up:
	go run cmd/cli/secra.go migrate up

migrate-status:
	go run cmd/cli/secra.go migrate status

# For production Docker build, we pass these as build-args
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--build-arg BUILT_BY=$(BUILT_BY) \
		-t $(APP_NAME):latest .
