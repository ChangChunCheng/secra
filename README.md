# SECRA (Security Resource Aggregator)

SECRA is a high-performance security vulnerability aggregation and subscription platform. It is designed to provide real-time, precise CVE monitoring and automated notification services for enterprises and developers.

> **Note:** This project is autonomously developed and maintained by a human developer commanding a **Gemini-based AI Agent**. Every line of code, architecture decision, and documentation was executed by the agent under strategic human direction.

## 🚀 Core Features (v0.0.2-alpha)

- **Modernized Frontend-Backend Separation**: Next.js SPA frontend with Go REST API backend.
- **Type-Safe API**: TypeScript types generated from Protobuf definitions.

- **Unified Monolith Architecture**: Both HTTP (8081) and gRPC (50051) servers run within a single container, sharing a unified resource pool for maximum efficiency.
- **Precise Asset Subscriptions**: Multi-dimensional subscription support for Vendors and Products.
- **Intelligent Notification Engine**:
  - **Batch Aggregation**: Prevents inbox flooding by consolidating large batches of new CVEs into a single, comprehensive alert email.
  - **Scheduled Digests**: Supports timezone-aware digests sent at user-defined intervals (e.g., daily at 08:00 AM local time).
- **Timezone Awareness**: Full support for localized timezones in both the Web UI and notification scheduling.
- **Deterministic Stability**: Global stable ordering across all asset and vulnerability lists to ensure a consistent user experience.

## 🛠️ Quick Start

### Prerequisites

- Docker & Docker Compose
- Make (optional, for convenience commands)
- [Buf CLI](https://buf.build/docs/installation) (only for local development)

**Note:** Generated code from Protobuf (`.proto` files) is **not** committed to git. It is automatically generated during Docker builds and should be generated locally for development.

### 1. Configuration
Clone the template and configure your SMTP/Database settings:
```bash
cp template.env .env
```

### 2. Launch System
Deploy the full stack using Docker Compose:
```bash
make docker-up
```

**Access Points:**
- **Frontend (Web UI)**: http://localhost (Port 80)
- **Backend API**: http://localhost:8081/api/v1/*
- **gRPC**: localhost:50051

## 🎯 System Initialization & Operations

### System Startup Behavior

When you start SECRA for the first time, the system automatically:

1. **Auto-Migration** (if `AUTO_MIGRATE=true` in `.env`, default: enabled)
   - Checks for pending database migrations
   - Applies all necessary schema updates
   - Subsequent startups skip migrations if database is up-to-date
   - Logs: `✅ All migrations already applied, skipping.`

2. **Auto-Create Default Admin** (if database has no users)
   - Creates admin account using credentials from `.env`:
     - Username: `SECRA_ADMIN_USER` (default: `admin`)
     - Password: `SECRA_ADMIN_PWD` (default: `admin`)
   - Logs: `✅ Default admin user 'admin' created successfully`
   - **⚠️ Change the default password in production!**

**No manual CLI commands required for first startup!** Just run:
```bash
docker compose up -d
```

### Default Admin User

| Field | Value | Source |
|-------|-------|--------|
| Username | `admin` | `SECRA_ADMIN_USER` in `.env` |
| Password | `admin` | `SECRA_ADMIN_PWD` in `.env` |
| Role | Admin | Auto-assigned |
| Auto-Created | ✅ Yes | On first startup if no users exist |

**Login Example:**
```bash
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}'
```

### CLI User Management

**Register a New User:**
```bash
docker compose exec server secra user register \
  --username john \
  --email john@example.com \
  --password secretpass
```

**Reset User Password:**
```bash
docker compose exec server secra user reset-password \
  -u john \
  -p newpassword
```

**List All Users:**
```bash
docker compose exec server secra user list
```

**Update User Role:**
```bash
docker compose exec server secra user update-role \
  --username john \
  --role admin
```

### Manual NVD Data Import

Import CVE data from the National Vulnerability Database:

**Import Recent CVEs (recommended for testing):**
```bash
docker compose exec server secra import nvd v2 \
  --start 2024-01-09 \
  --end 2024-01-10 \
  -f
```

**Import Historical Range:**
```bash
docker compose exec server secra import nvd v2 \
  --start 2023-01-01 \
  --end 2023-12-31
```

**Parameters:**
- `--start`: Start date in YYYY-MM-DD format (required)
- `--end`: End date in YYYY-MM-DD format (optional, defaults to today)
- `-f, --force`: Force re-import even if data exists
- `--apikey`: NVD API key for higher rate limits (optional)

**Note:** Without an API key, imports are rate-limited by NVD. Request a free API key at: https://nvd.nist.gov/developers/request-an-api-key

## 📈 Operations (Makefile)

**From project root:**

| Command | Description |
| :--- | :--- |
| `make help` | Show all available maintenance commands |
| `make backend-build` | Build backend binaries locally |
| `make backend-test` | Run backend unit and integration tests |
| `make frontend-build` | Build frontend static files |
| `make docker-up` | Build and launch all services in Docker |
| `make docker-down` | Stop and remove all containers |
| `make migrate-up` | Execute pending database migrations |
| `make clean` | Remove all build artifacts |

**Backend-specific commands:**
```bash
cd backend && make help  # See all backend commands
```

## 📁 Project Structure (Monorepo)

```
secra/
├── backend/                    # Go backend application
│   ├── cmd/                    # Application entry points
│   │   ├── server/             # HTTP + gRPC server
│   │   └── cli/                # CLI tools
│   ├── internal/               # Private backend code
│   │   ├── api/web/            # REST API handlers
│   │   ├── service/            # Business logic
│   │   ├── repo/               # Database layer
│   │   ├── model/              # Data models
│   │   └── ...
│   ├── migrations/             # Database migrations
│   ├── scripts/                # Maintenance scripts
│   ├── tests/                  # Backend tests
│   ├── go.mod                  # Go dependencies
│   ├── Dockerfile              # Backend container image
│   └── Makefile                # Backend build commands
│
├── frontend/                   # Next.js frontend application
│   ├── src/                    # Frontend source code
│   │   ├── app/                # Next.js pages
│   │   ├── components/         # React components
│   │   └── lib/                # State management
│   ├── tests/                  # Frontend tests (planned)
│   ├── Dockerfile              # Frontend container image
│   └── package.json
│
├── api/                        # Shared API definitions
│   └── v1/*.proto              # Protobuf schemas
│
├── docs/                       # Project documentation
├── tests/e2e/                  # End-to-end tests (planned)
├── docker-compose.yml          # Service orchestration
└── Makefile                    # Root-level commands
```

**📖 Detailed Documentation:**
- [Backend README](backend/README.md) - Backend architecture, API, CLI, testing
- [Frontend README](frontend/README.md) - Frontend setup, components, state management
- [API README](api/README.md) - Protobuf definitions, code generation, API specs

## 🧪 Testing Notifications
Trigger a batch import to verify subscription matching and email aggregation:
```bash
docker compose exec server secra import nvd v2 --start 2024-01-09 --end 2024-01-10 -f
```

## 📄 License
[LICENSE](LICENSE)
