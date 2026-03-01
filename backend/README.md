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

# Backup & Restore
secra backup create -o /path/to/backup.tar.gz
secra backup restore /path/to/backup.tar.gz

# Health check
secra health check --type http

# Version
secra version
secra version --raw  # Returns version string only
```

## 💾 Backup & Restore

### Overview

The backup/restore system uses **Parquet format** for efficient storage and supports **UUID v5 migration** for data consistency.

**Backed up data includes:**
- CVE Sources (`cve_sources`)
- Vendors & Products (`vendors`, `products`)
- CVEs & Relations (`cves`, `cve_products`)
- Users & Subscriptions (`users`, `subscriptions`)

### Creating Backups

**Recommended: Using CLI with auto-generated filename**

```bash
# Auto-generate filename (secra_<version>_<timestamp>.tar.gz) in container
docker compose exec server secra backup create -d /tmp

# Copy to host
docker cp secra-server:/tmp/secra_v0.0.2-alpha_20260301_143022.tar.gz ./backups/
```

**Alternative: Using mounted volume (recommended for production)**

```bash
# Add volume mount in docker-compose.yml:
# volumes:
#   - ./backups:/backups:rw

# Then backup directly to host directory
docker compose exec server secra backup create -d /backups
# Files appear immediately in ./backups/ on host
```

**Local Development (without Docker)**

```bash
cd backend
./bin/secra backup create -d ./backups
# Output: ./backups/secra_dev_20260301_143022.tar.gz
```

**Features:**
- ✅ Auto-generated filename with version and timestamp
- ✅ Auto-creates output directory if missing
- ✅ Custom path with `-o` or directory with `-d`

### Restoring from Backup

**Recommended: Using CLI (with volume mount)**

```bash
# If using volume mount (./backups:/backups in docker-compose.yml)
docker compose exec server secra backup restore /backups/secra_v0.0.2-alpha_20240315_143022.tar.gz
```

**Alternative: Copy to container then restore**

```bash
# Copy backup into container
docker cp ./backups/secra_v0.0.2-alpha_20240315_143022.tar.gz secra-server:/tmp/

# Restore from container filesystem
docker compose exec server secra backup restore /tmp/secra_v0.0.2-alpha_20240315_143022.tar.gz
```

**Local Development (without Docker)**

```bash
cd backend
./bin/secra backup restore ./backups/secra_dev_20260301_143022.tar.gz
```

**Restore Process:**
1. ✅ Validates backup file exists
2. ✅ Extracts and migrates UUIDs to v5 format
3. ✅ Imports data with conflict resolution (UPSERT)
4. ✅ Restores statistics and relationships

### Important Notes

**UUID v5 Migration:**
- Old random UUIDs are automatically migrated to deterministic UUID v5
- IDs are generated using: `uuid5(namespace, unique_key)`
  - CVE: `uuid5("cve", source_uid)` e.g., `CVE-2024-1234`
  - Vendor: `uuid5("vendor", name)` e.g., `microsoft`
  - Product: `uuid5("product", vendor:name)` e.g., `microsoft:windows_10`

**Conflict Handling:**
- Existing records with same ID are updated (UPSERT)
- No duplicate entries will be created
- Safe to restore multiple times

**Post-Restore:**
- Daily statistics (`daily_cve_counts`) are automatically recalculated
- Subscriptions remain intact with correct user associations

**Limitations:**
- `cve_products` relations may need re-sync for 100% accuracy
- Recommendation: Run NVD import after restore to rebuild perfect relations

### Backup Best Practices

1. **Regular Backups:**
   ```bash
   # Setup cron job (daily at 2 AM) - requires volume mount
   0 2 * * * docker compose -f /path/to/secra/docker-compose.yml exec -T server secra backup create -d /backups
   ```

2. **Retention Policy:**
   ```bash
   # Keep last 7 days, delete older
   find /path/to/backups -name "secra_*.tar.gz" -mtime +7 -delete
   ```

3. **Off-site Storage:**
   ```bash
   # Sync to remote storage
   rclone copy /path/to/backups remote:secra-backups
   ```

4. **Verify Backups:**
   ```bash
   # Test restore on staging environment regularly
   docker compose exec server secra backup restore /backups/$(ls -t /backups/secra_*.tar.gz | head -1)
   ```

### Troubleshooting

**Backup fails with "Container not running":**
```bash
# Ensure containers are up
docker compose ps
docker compose up -d
```

**Restore fails with permission error:**
```bash
# Check file permissions
ls -l ./backups/backup.tar.gz
chmod 644 ./backups/backup.tar.gz
```

**Large database backup is slow:**
- Parquet format is optimized but large datasets take time
- Consider streaming exports for databases > 1GB
- Compress with higher levels: `gzip -9`

**Relations not fully restored:**
```bash
# Re-run NVD import to rebuild perfect cve_products relations
docker compose exec server secra import nvd v2 --start 2024-01-01 -f
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
