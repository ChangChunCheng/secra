# Secra 專案用途與架構說明

## 1. 專案定位與功能

Secra 是一個用於管理與分析 CVE（Common Vulnerabilities and Exposures）漏洞資訊的模組化平台。目標在於：
    - 從多個來源（NVD、RedHat、JPCERT 等）同步最新漏洞資料  
    - 解析並標準化不同版本的 JSON 資料  
    - 儲存至 PostgreSQL，並提供 CRUD 存取  
    - 支援 CLI、自動排程（Cron）、gRPC 及 REST API  
    - 可擴充通知（Webhook、訂閱）與風險分析模組  

## 2. 主要目錄與模組

    - **cmd/**: CLI 與伺服器啟動程式  
    - **api/**: Protobuf 定義（`api/proto/v1/*.proto`）與生成程式碼（`api/gen/v1`）  
    - **internal/**: 核心業務邏輯  
    - **config**: 讀取環境變數與初始化設定  
    - **db**: 資料庫 Migration  
    - **model**: Bun ORM 資料表映射  
    - **fetcher**: 下載原始 JSON.gz  
    - **parser**: 解壓與轉換為內部結構  
    - **importer**: 寫入資料庫  
    - **repo**: 高階 CRUD 操作封裝  
    - **service**: 商業邏輯層（通知、風險分析）  
    - **server**: gRPC / REST API 路由與 Handler  
    - **storage**: 資料庫連線管理  
    - **migrations/**: SQL 建表腳本  
    - **Dockerfile & docker-compose.yml**: 容器化部署  
    - **run_nvd_cron.sh**: 定時任務執行腳本  
    - **template.env**: 環境變數範本  

## 3. 資料流程

參考 [ARCHITECTURE.md](ARCHITECTURE.md) 中的 Mermaid 圖示：
```{mermaid}
flowchart LR
  A[CLI or Cron] -->|觸發| B[Fetcher]
  B -->|下載 JSON gz| C[Parser]
  C -->|解析成內部結構| D[Importer]
  D -->|透過 Repo 寫入| E[(PostgreSQL)]
  E -->|提供存取| F[Server Layer]
  F -->|gRPC / REST API| G[Client 或其他模組]
  F -->|通知或 webhook| H[Service 擴充]
```

1. **Fetcher** 透過 HTTP 下載漏洞資料（JSON.gz）  
2. **Parser** 解壓、依版本(v1/v2)轉換成標準結構  
3. **Importer** 寫入 DB，藉由 **Repo** 層封裝 CRUD  
4. **Server Layer**（gRPC/REST）對外提供查詢、訂閱、匯入狀態等 API  
5. **Service 擴充** 可進行通知推播、Webhook 或風險分析  

## 4. 可擴充點

   - 新增其他漏洞來源 Fetcher/Parser  
   - 擴充 Service 加入郵件、Slack 通知  
   - 外部模組接入（Webhook、風險評估）  
   - 前端 UI 或 Dashboard  

以上即為 Secra 平台的用途與整體架構說明。
