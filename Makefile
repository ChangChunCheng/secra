# ============================================================================
# Metadata
# ============================================================================

VERSION ?= 0.1.0
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%d)

LDFLAGS := -X 'main.Version=$(VERSION)' \
           -X 'main.Commit=$(COMMIT)' \
           -X 'main.BuildDate=$(DATE)'

# ============================================================================
# Paths and output
# ============================================================================

PROTO_DIR := proto
OUT_DIR := api/gen/v1
PROTO_FILE := $(PROTO_DIR)/secra.proto
PROTO_GOOGLE := $(shell GOFLAGS=-mod=mod go list -f '{{ .Dir }}' -m google.golang.org/protobuf)

CLI_BIN := secra-cli
GRPC_BIN := secra-grpc
HTTP_BIN := secra-api

# ============================================================================
# Proto
# ============================================================================

proto:
	@mkdir -p $(OUT_DIR)
	protoc \
		-I$(PROTO_DIR) \
		-I$(PROTO_GOOGLE)/.. \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		$(PROTO_FILE)
	@echo "✅ Proto compiled."

# ============================================================================
# Build with metadata
# ============================================================================

build-cli:
	go build -ldflags "$(LDFLAGS)" -o bin/$(CLI_BIN) ./cmd/cli

build-grpc:
	go build -ldflags "$(LDFLAGS)" -o bin/$(GRPC_BIN) ./cmd/server/grpc_server

build-http:
	go build -ldflags "$(LDFLAGS)" -o bin/$(HTTP_BIN) ./cmd/server/http_server

build: build-cli build-grpc build-http

# ============================================================================
# Run
# ============================================================================

run-cli:
	go run ./cmd/cli

run-grpc:
	go run ./cmd/server/grpc_server

run-http:
	go run ./cmd/server/http_server

# ============================================================================
# Docker
# ============================================================================

docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE) -t secra .

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

dbshell:
	docker exec -it secra-db psql -U postgres -d secra

# ============================================================================
# Dev utils
# ============================================================================

mod-tidy:
	GOFLAGS=-mod=mod go mod tidy

fmt:
	go fmt ./...

lint:
	golangci-lint run

# ============================================================================
# Migrations & Import
# ============================================================================

migrate:
	go run cmd/cli/main.go migrate up

migrate-status:
	go run cmd/cli/main.go migrate status

import-nvd:
	go run cmd/cli/main.go import nvd --year=2025 --source=nvd-cve