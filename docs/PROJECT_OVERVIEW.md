# Secra 專案概觀

## 1. 宗旨

Secra 是一個全面的安全姿態管理工具，旨在幫助開發人員和安全專業人員主動追蹤和管理與其專案相關的軟體相依性的安全漏洞。

系統的核心功能是從公開的漏洞資料庫（如 NVD）中提取常見漏洞和暴露 (CVE) 資料，並允許使用者訂閱他們感興趣的特定軟體供應商或產品。當新的相關 CVE 發布時，系統可以通知使用者，從而實現快速的風險評估和緩解。

## 2. 核心架構

Secra 採用了現代化的多分層架構，主要由以下幾個組件構成：

### 2.1. 資料庫 (Database)

- **技術棧**: PostgreSQL
- **描述**: 作為資料持久層，儲存所有核心資料模型，包括使用者、CVE、供應商、產品和訂閱關係。
- **遷移管理**: 使用 `golang-migrate` 進行資料庫綱要 (Schema) 的版本控制和遷移。遷移腳本位於 `migrations/` 目錄下。

### 2.2. 後端伺服器 (Backend Server)

- **技術棧**: Go, gRPC, gRPC-Gateway
- **描述**: 專案的核心業務邏輯所在。它以 gRPC 伺服器的形式提供了一組強型別的 API，用於所有 CRUD（建立、讀取、更新、刪除）操作和業務邏輯。
- **API 定義**: gRPC 服務和訊息的定義位於 `api/v1/` 目錄下的 `.proto` 檔案中。
- **HTTP 閘道**: 透過 `gRPC-Gateway`，gRPC 服務也以 RESTful JSON 的形式暴露出來，為沒有原生 gRPC 支援的客戶端提供了便利的存取方式。

### 2.3. 命令列介面 (CLI)

- **技術棧**: Go, Cobra
- **描述**: `secra` CLI 是與後端伺服器互動的主要工具，提供了一個使用者友好的命令列介面來管理系統中的資源。
- **功能**: 使用者可以透過 CLI 進行註冊、登入、管理 CVE 來源、建立和管理訂閱等。
- **原始碼**: CLI 的入口點和命令定義位於 `cmd/cli/` 目錄下。

### 2.4. 前端應用 (Frontend)

- **技術棧**: Next.js 16, React 19, Redux Toolkit, TailwindCSS
- **描述**: 現代化的單頁應用程式 (SPA)，提供直觀的使用者介面。
- **特點**: 
  - 使用 Next.js App Router 進行路由管理
  - Redux Toolkit Query 用於 API 狀態管理
  - TypeScript 類型定義自動從 Protobuf 生成
  - 響應式設計，支援各種設備
- **原始碼**: 位於 `frontend/` 目錄下。

### 2.5. 資料匯入器 (Data Importer)

- **描述**: 一個整合在 CLI 中的專用工具，用於從外部來源（目前主要是美國國家漏洞資料庫 NVD）獲取和匯入 CVE 資料。
- **位置**: 相關邏輯位於 `cmd/cli/import/` 和 `internal/fetcher/`、`internal/parser/`、`internal/importer/` 等目錄中。

## 3. 目錄結構

```
/
├── api/            # gRPC 和 Protobuf API 定義
├── cmd/            # 應用程式入口點 (伺服器和 CLI)
│   ├── server/     # 後端 gRPC/HTTP 伺服器
│   └── cli/        # 命令列介面
├── frontend/       # Next.js 前端應用程式
│   ├── src/        # 源代碼
│   │   ├── app/    # Next.js 頁面和路由
│   │   ├── components/ # React 組件
│   │   └── lib/    # 工具函式和 Redux store
│   └── public/     # 靜態資源
├── internal/       # 私有的應用程式和函式庫程式碼
│   ├── api/        # REST API handlers
│   ├── auth/       # 認證邏輯 (密碼雜湊)
│   ├── config/     # 設定管理
│   ├── model/      # 核心業務模型 (Structs)
│   ├── repo/       # 資料庫倉儲層 (CRUD)
│   ├── service/    # 業務邏輯服務層
│   └── storage/    # 資料庫連線管理
├── migrations/     # 資料庫遷移腳本
└── docs/           # 專案文件
```

## 4. 工作流程範例

1.  **管理員**: 使用 `secra import nvd` CLI 命令從 NVD 匯入最新的 CVE 資料。
2.  **使用者**: 透過 `secra register` CLI 命令建立帳戶。
3.  **使用者**: 使用 `secra subscribe product <product_name>` 命令訂閱一個他們正在使用的軟體產品。
4.  **系統**: 在後台，Secra 將此訂閱與現有的 CVE 資料進行關聯。
5.  **通知**: 當新的 CVE 被匯入並與使用者訂閱的產品匹配時，系統（在未來的版本中）可以觸發通知。
