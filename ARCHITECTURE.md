# Secra 架構圖

```{mermaid}
flowchart TB
  A[CLI or Cron] -->|觸發| B[Fetcher]
  B -->|下載 JSON gz| C[Parser]
  C -->|解析成內部結構| D[Importer]
  D -->|透過 Repo 寫入| E[(PostgreSQL)]
  E -->|提供存取| F[Server Layer]
  F -->|gRPC / REST API| G[Client 或其他模組]
  F -->|通知或 webhook| H[Service 擴充]
```
