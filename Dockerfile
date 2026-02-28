# Stage 0: Generate code from Protobuf
FROM bufbuild/buf:latest AS generate

WORKDIR /app
COPY api ./api
RUN cd api && buf generate --template buf.gen.yaml

# Stage 1: Build binaries
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies (golang:alpine has CA certs needed for this)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy generated code from Stage 0
COPY --from=generate /app/api/gen ./api/gen

# Build static binaries
# -s -w removes symbol table and debug information to reduce size
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o secra-grpc ./cmd/server/grpc.go && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o secra-http ./cmd/server/http.go && \
    CGO_ENABLED=0 go build -ldflags="-s -w" -o secra ./cmd/cli/secra.go

# Stage 2: Final Runtime Image (The Professional Choice)
# gcr.io/distroless/static contains CA certificates, /etc/passwd, and tzdata
# but NO shell, NO package manager - ideal for static Go binaries.
FROM gcr.io/distroless/static-debian12

WORKDIR /app

# Copy binaries and assets
COPY --from=builder /app/secra-grpc .
COPY --from=builder /app/secra-http .
COPY --from=builder /app/secra /usr/local/bin/secra
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations

# Expose ports
EXPOSE 50051 8081

# Default command
CMD ["./secra-http"]
