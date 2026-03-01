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

**Key Configuration Options:**

<details>
<summary>🗄️ <strong>Database Configuration</strong></summary>

```env
# PostgreSQL Database Settings
POSTGRES_USER=postgres          # Database username
POSTGRES_PASSWORD=postgres      # Database password (⚠️ Change in production!)
POSTGRES_DB=secra              # Database name
POSTGRES_HOST=db               # Docker: use 'db' | Local dev: use 'localhost'
POSTGRES_PORT=5432             # Internal port (no external mapping needed)
```

**Notes:**
- Connection string is auto-constructed from these variables
- Database port is **NOT exposed externally** by default (secure by design)
- All services communicate via internal Docker network
- For local development debugging, uncomment port mapping in `docker-compose.yml`

**Debug Access (Optional):**
```yaml
# Uncomment in docker-compose.yml if needed:
# db:
#   ports:
#     - "5432:5432"
```

</details>

<details>
<summary>📧 <strong>Email Notification Settings</strong></summary>

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASS=your-app-password
SMTP_FROM=noreply@secra.local
SMTP_ENCRYPTION=STARTTLS     # Options: SSL, STARTTLS, NONE
```

</details>

<details>
<summary>⏰ <strong>CVE Import Scheduler</strong></summary>

```env
IMPORT_ENABLED=true           # Enable automatic CVE imports
IMPORT_SCHEDULE=0 0 * * * *   # Cron format with seconds (default: hourly)
```

**Schedule Examples:**
- `0 0 * * * *` - Every hour at :00 minutes (default)
- `0 0 0 * * *` - Daily at midnight
- `0 0 */6 * * *` - Every 6 hours
- `0 30 2 * * *` - Daily at 2:30 AM

</details>

### 2. Launch System
Deploy the full stack using Docker Compose:
```bash
make docker-up
```

**Access Points:**
- **Frontend (Web UI)**: http://localhost (Port 80) - **唯一對外進入點**
- **Backend API**: 通過 Frontend 反向代理訪問 `/api/v1/*`
- **gRPC**: 內部服務通訊（不對外開放）

**架構說明：**
- ✅ 所有對外請求都通過 Frontend (Nginx) 統一入口
- ✅ Backend 和 Database 僅在 Docker 內部網路通訊
- ✅ 生產環境無需暴露內部服務端口
- 🔧 開發調試需要時，可在 `docker-compose.yml` 中取消註釋相關 ports

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

### Automated CVE Import Scheduler

SECRA includes an **automated scheduler** that periodically imports new CVEs from configured sources (currently NVD), eliminating the need for manual CLI imports in production.

**Key Features:**
- ✅ **Automatic Scheduled Imports**: Runs at configurable intervals (default: hourly)
- ✅ **Auto-Backfill on Startup**: Automatically imports missing CVEs since last successful run
- ✅ **Import History Tracking**: All import jobs logged in `import_jobs` table
- ✅ **Extensible Architecture**: Easy to add new CVE sources (OSV, GitHub Advisories, etc.)
- ✅ **Rate Limiting**: Respects API rate limits with daily chunking

**Configuration:**

The scheduler is configured via environment variables in `.env`:

```env
# CVE Import Scheduler Settings
IMPORT_ENABLED=true                    # Enable/disable scheduler (default: true)
IMPORT_SCHEDULE=0 0 * * * *            # Cron format with seconds (default: every hour)
```

**Cron Format:** `seconds minutes hours day month weekday`

**Common Schedule Examples:**

| Schedule | Cron Expression | Description |
|----------|-----------------|-------------|
| Every hour | `0 0 * * * *` | Runs at :00 of every hour |
| Every 6 hours | `0 0 */6 * * *` | Runs at 00:00, 06:00, 12:00, 18:00 |
| Daily at 2:30 AM | `0 30 2 * * *` | Runs once per day at 02:30 |
| Twice daily | `0 0 9,21 * * *` | Runs at 09:00 and 21:00 |

**Scheduler Behavior:**

1. **On Server Startup:**
   - Checks last successful import date for each source
   - Automatically imports all missing days since then (backfill)
   - Example: If server was offline for 3 days, it imports those 3 days' CVEs on startup

2. **During Scheduled Runs:**
   - Imports CVEs for the current day
   - Records job status in `import_jobs` table
   - Continues on next schedule even if a job fails

**Disable Scheduler:**

To disable automatic imports (for testing or manual control):
```env
IMPORT_ENABLED=false
```

Then use manual CLI imports as needed.

**Monitor Import History:**

View import job history via database query:
```sql
SELECT * FROM import_jobs 
ORDER BY start_time DESC 
LIMIT 10;
```

Or check server logs:
```bash
docker compose logs -f server | grep "NVD"
```

## 💾 Backup & Restore

**Quick Start:**

```bash
# Create backup (auto-generates: secra_<version>_<timestamp>.tar.gz)
docker compose exec server secra backup create -d /tmp
docker cp secra-server:/tmp/secra_v0.0.2-alpha_20240315_143022.tar.gz ./backups/

# Restore from backup (copy file to container first)
docker cp ./backups/secra_v0.0.2-alpha_20240315_143022.tar.gz secra-server:/tmp/
docker compose exec server secra backup restore /tmp/secra_v0.0.2-alpha_20240315_143022.tar.gz
```

**Recommended: Use Volume Mount for Direct Access**

Add to `docker-compose.yml`:
```yaml
services:
  server:
    volumes:
      - ./backups:/backups:rw
```

Then backup and restore directly:
```bash
# Backup - files appear immediately in ./backups/
docker compose exec server secra backup create -d /backups

# Restore - access files directly
docker compose exec server secra backup restore /backups/secra_v0.0.2-alpha_20240315_143022.tar.gz
```

**What's Backed Up:**
- CVE data: Sources, CVEs, references, weaknesses, products (17 tables total)
- Security data: Vendors and products catalog
- User data: Accounts, roles, OAuth accounts, subscriptions
- System data: Severity levels, target types, daily statistics
- Format: Parquet-based compressed tar.gz for efficiency

**Key Features:**
- ✅ Auto-generated filename with version and timestamp
- ✅ Automatic UUID v5 migration for data consistency
- ✅ Conflict-safe restore (no duplicates, uses UPSERT)
- ✅ Complete data integrity with all relationships
- ✅ Pure CLI - no shell script dependencies

**⚠️ Important Notes:**
- Database schema must match backup version (run migrations first if needed)
- Large databases may take several minutes to backup/restore
- Regular backup schedule recommended (see backend README for cron setup)

📖 **For detailed documentation, troubleshooting, and best practices, see [backend/README.md - Backup & Restore](backend/README.md#-backup--restore)**

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
