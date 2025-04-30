ENV_FILE := .env
ifneq ("$(wildcard $(ENV_FILE))","")
	include $(ENV_FILE)
	export
endif

export PATH=$PATH:$(go env GOPATH)/bin


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

CLI_BIN := secra-cli
GRPC_BIN := secra-grpc
HTTP_BIN := secra-api

# ============================================================================
# Environment
# ============================================================================
YEAR ?= $(shell date -u +%Y)
NVD_API_KEY ?= 1234567890
START ?= $$(date -u -d '72 hours ago' +%Y-%m-%d)
END ?= $$(date -u -d '48 hours ago' +%Y-%m-%d)

# ============================================================================
# Proto
# ============================================================================

install-proto-tools:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	@echo "✅ All protoc plugins installed."

PROTO_DIR=api/proto/v1
OUT_DIR=api/gen/v1
PROTO_SRCS=$(shell find $(PROTO_DIR) -name "*.proto")
PROTO_GOOGLE=$(shell go list -m -f '{{ .Dir }}' github.com/grpc-ecosystem/grpc-gateway/v2)
PROTOBUF_GOOGLE=$(shell go list -f '{{ .Dir }}' google.golang.org/protobuf)
GRPC_GATEWAY_GOOGLE=$(shell go list -f '{{ .Dir }}' github.com/grpc-ecosystem/grpc-gateway/v2)
proto: $(PROTO_SRCS)
	@mkdir -p $(OUT_DIR)
	protoc \
		-I$(PROTO_DIR) \
		-I$(GRPC_GATEWAY_GOOGLE)/../.. \
		-I$(PROTOBUF_GOOGLE)/.. \
		--go_out=$(OUT_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUT_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(OUT_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_SRCS)


swagger: $(PROTO_SRCS)
	@mkdir -p $(OUT_DIR)
	protoc \
		-I$(PROTO_DIR) \
		-I$(PROTO_GOOGLE)/../.. \
		-I$(PROTOBUF_GOOGLE)/.. \
		--openapiv2_out=$(OUT_DIR) --openapiv2_opt=logtostderr=true \
		$(PROTO_SRCS)
	@echo "✅ Swagger/OpenAPI generated."

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
	go run cmd/cli/secra.go migrate up

migrate-status:
	go run cmd/cli/secra.go migrate status

import-nvd-v1-recent:
	go run cmd/cli/secra.go import nvd v1 --recent=true

import-nvd-v1:
	go run cmd/cli/secra.go import nvd v1 --recent=true --modified=true --year=$(YEAR)

# 匯入 NVD v2 (新版) → 需要傳入起始日（YYYY-MM-DD），結束日可選
import-nvd-v2:
ifneq ($(strip $(START)),)
	go run cmd/cli/secra.go import nvd v2 --start=$(START) $(if $(END),--end=$(END)) $(if $(APIKEY),--apikey=$(APIKEY))
else
	@echo "❌ 必須指定 START 日期 (YYYY-MM-DD)。例如：make import-nvd-v2 START=2025-01-01 END=2025-01-31"
	@exit 1
endif