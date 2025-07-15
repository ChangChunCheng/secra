# Secra

Secra 是一個模組化的 CVE 漏洞資料平台，支援多來源同步、自定義通報、訂閱通知與可擴充模組架構。

## 架構特色

- 使用 Golang + Bun ORM 開發
- 支援 PostgreSQL 儲存
- 支援 CLI / gRPC / REST API 操作
- 可整合 NVD、RedHat、JPCERT 等來源
- 擴充性強，可接 webhook 或風險分析模組

## 快速開始

```{bash}
git clone git@gitlab.com:jacky850509/secra.git
cd secra
go run cmd/cli/secra.go
```

---

## 目錄結構

```{bash}
secra/
├── cmd/
│   ├── cli/         # CLI command 定義
│   │   ├── secra.go
│   │   └── root/
│   │       └── root.go
│   └── cron/        # 未來排程用 CLI 入口
├── internal/
│   ├── model/       # 資料表 ORM 定義
## 啟動 gRPC 伺服器

在使用 CLI 的 `create-vendor` 等 gRPC 相關命令前，需先啟動 gRPC 伺服器：

```bash
go run cmd/server/grpc_server/register.go
```
│   ├── db/          # 資料庫初始化
│   ├── fetcher/     # 抓取來源資料(如 NVD)
│   ├── parser/      # 解壓與解析 JSON
│   ├── importer/    # 將資料寫入資料庫
│   └── api/         # bunrouter REST API(未來)
├── migrations/      # SQL 建表
├── data/            # 原始 JSON.gz 儲存路徑
└── README.md
```

## 使用者註冊與 OAuth2 登入

1. 重新產生 gRPC Stub  
   ```bash
   protoc --go_out=api/gen --go-grpc_out=api/gen api/proto/v1/user.proto
   ```

2. 啟動 gRPC 伺服器  
   ```bash
   go run cmd/server/grpc_server/main.go
   ```

3. 測試 REST 端點  
   - 系統註冊  
     ```bash
     curl -X POST http://localhost:8080/v1/users/register \
       -H "Content-Type: application/json" \
       -d '{"username":"alice","email":"alice@example.com","password_hash":"<hashed>"}'
     ```  
   - OAuth2 登入  
     ```bash
     curl -X POST http://localhost:8080/v1/oauth/login \
       -H "Content-Type: application/json" \
       -d '{"provider":"github","provider_user_id":"12345","email":"alice@example.com"}'
     ```  
