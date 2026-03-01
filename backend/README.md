# SECRA Backend

Go-based backend server for SECRA (Security Resource Aggregator).

## 🏗️ Architecture

- **HTTP Server**: REST API on port 8080 (mapped to 8081 externally)
- **gRPC Server**: gRPC services on port 50051
- **Database**: PostgreSQL 17

## 📦 Project Structure

```
backend/
├── cmd/
│   ├── server/        # Main server application (HTTP + gRPC)
│   └── cli/           # CLI tools (secra command)
├── internal/
│   ├── api/web/       # REST API handlers
│   ├── service/       # Business logic layer
│   ├── repo/          # Database repository layer
│   ├── model/         # Data models
│   ├── auth/          # Authentication utilities
│   ├── config/        # Configuration management
│   ├── storage/       # Database connection
│   ├── fetcher/       # External data fetchers (NVD)
│   ├── parser/        # Data parsers
│   └── importer/      # Data importers
├── migrations/        # Database migration scripts
├── scripts/           # Utility scripts (backup, restore)
├── tests/             # Test suites
│   ├── unit/          # Unit tests
│   ├── integration/   # Integration tests
│   └── fixtures/      # Test data
├── api/               # Symlink to ../api (Protobuf definitions)
├── go.mod             # Go module definition
├── Dockerfile         # Container image definition
└── Makefile           # Build and deployment commands
```

## 🛠️ Development

### Prerequisites

- Go 1.25+
- PostgreSQL 17
- [Buf CLI](https://buf.build/docs/installation) (for protobuf generation)

**Important:** Generated Protobuf code (`api/gen/`) is **not** committed to git and must be generated before building.

### Local Build

```bash
# Generate Protobuf code first
cd ../api && buf generate && cd ../backend

# Build binaries
make build
```

Binaries will be output to `../bin/`:
- `secra-server` - Main server
- `secra` - CLI tool

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires database)
make test-integration

# Coverage report
make test-coverage
```

### Database Migrations

```bash
# Apply migrations
make migrate-up

# Check migration status
make migrate-status
```

## 🐳 Docker

### Start Services

```bash
make docker-up
```

### Stop Services

```bash
make docker-down
```

### Execute Commands in Container

```bash
# Run CLI commands
docker compose exec server secra user list

# Import CVE data
docker compose exec server secra import nvd v2 --start 2024-01-01 --end 2024-01-31
```

## 📚 API Documentation

### REST API

Base URL: `http://localhost:8081/api/v1`

Key endpoints:
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/me` - Get current user
- `GET /api/v1/cves` - List CVEs
- `GET /api/v1/cves/:id` - Get CVE details
- `GET /api/v1/vendors` - List vendors
- `GET /api/v1/products` - List products
- `GET /api/v1/my/dashboard` - User dashboard
- `POST /api/v1/subscriptions` - Create subscription
- `DELETE /api/v1/subscriptions?id=:id` - Delete subscription

### gRPC API

Port: `50051`

Protobuf definitions: `../api/v1/*.proto`

## 🔧 Configuration

Configuration is loaded from environment variables (see `template.env`):

```env
# Server
GRPC_PORT=:50051
HTTP_PORT=:8080

# Database
POSTGRES_DSN=postgres://user:pass@host:5432/dbname?sslmode=disable

# JWT
JWT_SECRET=your-secret-key

# NVD API
NVD_API_KEY=your-nvd-api-key

# SMTP (for notifications)
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=user
SMTP_PASS=password
```

## 📝 CLI Commands

```bash
# User management
secra user create -u username -e email@example.com -p password
secra user list
secra user reset-password --username user --password newpass
secra user update-role --username user --role admin
secra user delete --username user

# Database
secra migrate up
secra migrate status

# Data import
secra import nvd v2 --start 2024-01-01 --end 2024-01-31

# Health check
secra health check --type http
```

## 🧪 Testing Strategy

### Unit Tests
Test individual functions and methods in isolation.

```bash
go test -v -short ./tests/unit/...
```

### Integration Tests
Test API endpoints with real database (testcontainers).

```bash
go test -v -run Integration ./tests/integration/...
```

### Test Coverage
Generate HTML coverage report.

```bash
make test-coverage
open coverage.html
```

## 🚀 Deployment

### Production Build

```bash
# Build optimized binaries
CGO_ENABLED=0 go build -ldflags="-s -w" -o secra-server ./cmd/server/main.go
CGO_ENABLED=0 go build -ldflags="-s -w" -o secra ./cmd/cli/secra.go
```

### Docker Image

```bash
docker build -f Dockerfile -t secra-backend:latest ..
```

## 📄 License

See [LICENSE](../LICENSE)
