# Stage 0: Buf generation
FROM bufbuild/buf:latest AS generate
WORKDIR /app
COPY api ./api
RUN cd api && buf generate --template buf.gen.yaml

# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

# Arguments injected from host (Makefile / Docker Compose)
ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=none
ARG BUILT_BY=unknown

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=generate /app/api/gen ./api/gen

# Compile Consolidated Server
RUN export GO_OS=$(go env GOOS) && \
    export GO_ARCH=$(go env GOARCH) && \
    export PKG="gitlab.com/jacky850509/secra/internal/config" && \
    export LDFLAGS="-X ${PKG}.Version=${VERSION} -X ${PKG}.BuildDate=${BUILD_DATE} -X ${PKG}.Commit=${GIT_COMMIT} -X ${PKG}.BuiltBy=${BUILT_BY} -X ${PKG}.OS=${GO_OS} -X ${PKG}.Arch=${GO_ARCH} -s -w" && \
    CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o secra-server ./cmd/server/main.go && \
    CGO_ENABLED=0 go build -ldflags="${LDFLAGS}" -o secra ./cmd/cli/secra.go

# Stage 2: Final runtime image (Distroless for security)
FROM gcr.io/distroless/static-debian12:latest
WORKDIR /app
COPY --from=builder /app/secra-server .
COPY --from=builder /app/secra /usr/local/bin/secra
COPY --from=builder /app/web ./web
COPY --from=builder /app/migrations ./migrations

# Expose both Ports
EXPOSE 8081 50051
ENTRYPOINT ["./secra-server"]
