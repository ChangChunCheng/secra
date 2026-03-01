# SECRA - Vulnerability Intelligence Platform

> **Important Notice:** This entire project was developed and maintained exclusively by **AI Agents**.

SECRA is a high-performance, containerized CVE vulnerability platform featuring smart NVD synchronization, a Cyberpunk-styled Web UI, and columnar backup/restore capabilities.

---

## 🛠 Deployment Guide

### Prerequisites
- **Git**: Used for version tracking and build labeling.
- **Docker & Docker Compose**: For containerized execution.
- **Make**: To manage the unified build and deployment workflow.

### 1. Environment Setup
Copy the template and configure your environment variables:
```bash
cp template.env .env
```
Key variables in `.env`:
- `NVD_API_KEY`: (Optional) Highly recommended to increase rate limits (from 6s per request to 1s).
- `JWT_SECRET`: Used for user session authentication.
- `POSTGRES_DSN`: Should match the settings in `docker-compose.yml`.

### 2. Launch the System
Use the Makefile to ensure Git metadata is correctly injected into the containers:
```bash
make docker-up
```
*Note: This command captures your host's Git tags, commit hash, and hostname to bake them into the `secra` binary.*

### 3. Initialize Database
Once the containers are running, apply the schema migrations:
```bash
docker compose exec web secra migrate up
```

### 4. Fetch Initial Data
Pull vulnerability data from NVD (example for Jan 2024):
```bash
docker compose exec web secra import nvd v2 --start 2024-01-01 --end 2024-01-31
```

---

## 📊 Management Commands

### Operations
| Command | Description |
|---------|-------------|
| `make docker-up` | Build and start all services with Git metadata |
| `make docker-down` | Stop all services |
| `./backup.sh <dir>` | Create a Parquet-based backup with auto-versioning |
| `./restore.sh <file>` | Restore data from a backup (supports ID migration) |
| `docker compose exec web secra version` | Check detailed build & version info |

### Accessing the UI
The Cyberpunk Dashboard is available at: **http://localhost:8081**

---

## 🛠 Development Guidelines
For detailed information on branching, commit messages, and release cycles, please refer to:
**[Git Workflow & Standards](docs/GIT_WORKFLOW.md)**

---

## 🛡 License
Licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See [LICENSE](LICENSE) for details.
