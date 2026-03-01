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

### 3. Initialize Database
Run migrations to set up the latest schema:
```bash
make migrate-up
```

### 4. Create Admin Account
```bash
docker compose exec server secra user create -u admin -e admin@secra.local -p yourpassword --admin
```

## 📈 Operations (Makefile)

| Command | Description |
| :--- | :--- |
| `make help` | Show all available maintenance commands |
| `make build` | Build binaries locally for testing |
| `make docker-up` | Build and launch the Monolith server in Docker |
| `make migrate-up` | Execute pending database migrations |
| `make backup` | Execute full system backup (Optional: `OUT=./path`) |
| `make restore` | Restore system from backup (`FILE=path/to/file`) |

## 📁 Project Structure

- `/cmd/server`: Unified server entry point (Consolidated HTTP + gRPC).
- `/cmd/cli`: System management and data ingestion tools (NVD Fetcher).
- `/internal/service`: Core business logic (Notifications, Subscriptions, Auth).
- `/internal/api/web`: RESTful API handlers for frontend communication.
- `/frontend`: Modern Next.js frontend with React, Redux Toolkit, and TailwindCSS.
- `/api`: Protobuf definitions for gRPC services and TypeScript generation.
- `/scripts`: Automated maintenance and DevOps scripts.

## 🧪 Testing Notifications
Trigger a batch import to verify subscription matching and email aggregation:
```bash
docker compose exec server secra import nvd v2 --start 2024-01-09 --end 2024-01-10 -f
```

## 📄 License
[LICENSE](LICENSE)
