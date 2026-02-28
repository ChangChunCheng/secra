# SECRA - Vulnerability Intelligence Platform

> **Important Notice:** This entire project was developed and maintained exclusively by **AI Agents**.

SECRA is a high-performance, containerized CVE vulnerability platform featuring smart NVD synchronization, a Cyberpunk-styled Web UI, and columnar backup/restore capabilities.

## Key Features
- **Deterministic ID (UUID v5)**: Ensuring cross-environment data consistency.
- **Smart NVD Sync**: Greedy interval merging to minimize API calls and respect rate limits.
- **Cyberpunk Web UI**: Dark-themed, high-density dashboard with real-time analytics.
- **Columnar Backup**: Native Parquet-based backup and restore logic.
- **Distroless Runtime**: Secure, minimal Docker images.

## Quick Start
```bash
docker compose up -d
docker compose exec web secra migrate up
docker compose exec web secra import nvd v2 --start 2024-01-01
```

## Management
- **Backup**: `./backup.sh ./backups`
- **Restore**: `./restore.sh ./backups/filename.tar.gz`

## License
Licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See [LICENSE](LICENSE) for details.
