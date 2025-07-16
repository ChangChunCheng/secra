# Stage 1: Build binary
FROM golang:1.24.5 as builder

WORKDIR /app

COPY . .

ARG VERSION
ARG COMMIT
ARG DATE

RUN go mod download

# Build CLI
RUN go build -ldflags="-X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.BuildDate=${DATE}'" \
    -o bin/secra-cli ./cmd/cli

# Build gRPC server
RUN go build -ldflags="-X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.BuildDate=${DATE}'" \
    -o bin/secra-grpc ./cmd/server/grpc_server

# Build HTTP server
RUN go build -ldflags="-X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.BuildDate=${DATE}'" \
    -o bin/secra-api ./cmd/server/http_server

# Stage 2: Runtime
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/bin/secra-cli ./secra-cli
COPY --from=builder /app/bin/secra-grpc ./secra-grpc
COPY --from=builder /app/bin/secra-api ./secra-api

CMD ["./secra-cli"]