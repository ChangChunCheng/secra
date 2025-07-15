SHELL := /bin/bash
.SHELLFLAGS := -lc

# ----------------------------------------------------------------------------
# Load environment file if exists
# ----------------------------------------------------------------------------
ENV_FILE := .env
ifneq ("$(wildcard $(ENV_FILE))","")
	include $(ENV_FILE)
endif

# ----------------------------------------------------------------------------
# Go version enforcement via gvm
# ----------------------------------------------------------------------------
GO_REQUIRED := go1.24.5
GO_CURRENT := $(shell go version | awk '{print $$3}')
ifneq ($(GO_CURRENT),$(GO_REQUIRED))
	$(warning Current Go version is $(GO_CURRENT); switching to $(GO_REQUIRED) via gvm)
	# Load gvm and use required Go
	# Requires gvm installed and configured
	-@source ~/.gvm/scripts/gvm && gvm use $(GO_REQUIRED)
endif

# ----------------------------------------------------------------------------
# Build metadata (injected via -ldflags)
# ----------------------------------------------------------------------------
VERSION   ?= 0.1.0
COMMIT    := $(shell git rev-parse --short HEAD)
BUILD_DATE:= $(shell date -u +%Y-%m-%d)
LDFLAGS   := -X 'main.Version=$(VERSION)' \
             -X 'main.Commit=$(COMMIT)' \
             -X 'main.BuildDate=$(BUILD_DATE)'

# ----------------------------------------------------------------------------
# Protobuf code generation
# ----------------------------------------------------------------------------
BUF_TEMPLATE := api/buf.gen.yaml
PROTO_DIR    := api/proto/v1
GEN_OUT      := api/gen/v1

.PHONY: buf-gen
buf-gen:
	cd api && buf generate --template $(notdir $(BUF_TEMPLATE))

.PHONY: proto-gen
proto-gen: buf-gen
	@echo "Generating Go code with protoc"
	@mkdir -p $(GEN_OUT)
	protoc \
	  -I$(PROTO_DIR) \
	  -I$(shell go list -m -f '{{ .Dir }}/api/proto/v1' gitlab.com/jacky850509/secra) \
	  --go_out=$(GEN_OUT) --go_opt=paths=source_relative \
	  --go-grpc_out=$(GEN_OUT) --go-grpc_opt=paths=source_relative \
	  $(PROTO_DIR)/*.proto

# ----------------------------------------------------------------------------
# Build commands
# ----------------------------------------------------------------------------
.PHONY: build-cli build-grpc build-http build
build-cli:
	go build -ldflags "$(LDFLAGS)" -o bin/secra-cli ./cmd/cli

build-grpc:
	go build -ldflags "$(LDFLAGS)" -o bin/secra-grpc ./cmd/server/grpc_server

build-http:
	go build -ldflags "$(LDFLAGS)" -o bin/secra-api ./cmd/server/http_server

build: build-cli build-grpc build-http

# ----------------------------------------------------------------------------
# Run commands
# ----------------------------------------------------------------------------
.PHONY: run-cli run-grpc run-http
run-cli:
	go run ./cmd/cli

run-grpc:
	go run ./cmd/server/grpc_server

run-http:
	go run ./cmd/server/http_server

# ----------------------------------------------------------------------------
# Docker utilities
# ----------------------------------------------------------------------------
.PHONY: docker-build up down logs dbshell
docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(BUILD_DATE) -t secra .

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f

dbshell:
	docker exec -it secra-db psql -U postgres -d secra

# ----------------------------------------------------------------------------
# Development utilities
# ----------------------------------------------------------------------------
.PHONY: fmt lint mod-tidy
fmt:
	go fmt ./...

lint:
	golangci-lint run

mod-tidy:
	GOFLAGS=-mod=mod go mod tidy

# ----------------------------------------------------------------------------
# Migrations
# ----------------------------------------------------------------------------
.PHONY: migrate migrate-status
migrate:
	go run cmd/cli/secra.go migrate up

migrate-status:
	go run cmd/cli/secra.go migrate status