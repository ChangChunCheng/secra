# Secra: Vulnerability Intelligence Platform

Secra 是一個專業級的 CVE 漏洞情報監控平台，支援 NVD 自動同步、多維度訂閱通知、以及高性能的資安數據可視化儀表板。

## 🚀 核心功能

*   **智能 NVD 同步**：支援 NVD API v2 的高效同步，具備「區間感應」斷點續傳機制與自動節流重試。
*   **科技感 Web UI**：採用 Cyberpunk 深色風格設計，提供全域與個人化的監控儀表板。
*   **複合式搜尋**：支援關鍵字、廠商、產品及日期區間的多條件交叉查詢。
*   **資產訂閱系統**：可針對特定廠商 (Vendor) 或產品 (Product) 進行情報追蹤。
*   **原生備份還原**：基於 Parquet 列式儲存格式的高壓縮比備份方案。
*   **強健的架構**：整合 gRPC 與 REST API，支援 Docker 一鍵部署與健康監控。

## 🛠 快速開始

### 1. 啟動環境 (Docker 一鍵啟動)
確保您已安裝 Docker 與 Docker Compose，然後執行：
```bash
docker compose up -d --build
```
啟動後可訪問：[http://localhost:8081](http://localhost:8081)

### 2. 初始化資料庫
在容器啟動後執行資料庫遷移：
```bash
docker compose exec web secra migrate up
```

### 3. 建立管理員帳號
```bash
docker compose exec web secra user register --username admin --email admin@secra.io --password adminpassword
```

## 📖 CLI 指令說明

Secra 提供強大的 CLI 工具，透過 `docker compose exec web secra` 即可執行：

### 情報匯入 (NVD v2)
```bash
# 自動補齊缺失區間
docker compose exec web secra import nvd v2 --start 2025-01-01

# 強制重新匯入特定日期
docker compose exec web secra import nvd v2 --start 2026-02-28 -f
```

### 系統備份與還原 (Parquet 格式)
```bash
# 建立備份
docker compose exec web secra backup create -o /app/backup.tar.gz

# 執行還原
docker compose exec web secra backup restore /app/backup.tar.gz
```

### 健康檢查
```bash
docker compose exec web secra health check --type=db
```

## 🏗 技術架構

*   **Backend**: Go 1.25, Bun ORM, gRPC, gRPC-Gateway.
*   **Database**: PostgreSQL 15.
*   **Frontend**: Go html/template, Vanilla CSS, Chart.js.
*   **Storage**: Parquet (for backups).
*   **Deployment**: Docker (Distroless runtime for security).

## 🧪 測試

系統包含完整的 E2E 測試，確保 Web UI 功能正常：
```bash
cd tests/e2e
uv run pytest test_web_ui.py
```

---
*SECRA System &copy; 2026. Secure Vulnerability Intelligence Platform.*
