# CLI 子功能測試指令範例

以下示範如何使用各子指令進行操作，需以 `go run cmd/cli/secra.go` 作為前綴執行，假設使用 gvm 安裝的 Go 執行環境已正確設定。

## 1. 使用者註冊 (user register)

```bash
go run cmd/cli/secra.go user register \
  --username alice \
  --email alice@example.com \
  --password "securepassword"
```

輸出範例：

```bash
Registered user: alice (id=<user-uuid>)
```

## 2. 使用者登入並取得 JWT Token (user login)

```bash
go run cmd/cli/secra.go user login \
  --username alice \
  --password "securepassword"
```

輸出範例：

```bash
6f00ce7c-6b18-4b4e-a861-b58c4f221d04
```

## 3. 新增 CVE 資源 (resource create-cve-resource)

```bash
go run cmd/cli/secra.go resource create-cve-resource \
  --name "VendorX Resource" \
  --url "https://vendorx.com/feed"
```

輸出範例：

```bash
CVE Source created: ID=<source-uuid> Name=VendorX Resource URL=https://vendorx.com/feed
```

## 4. 新增 CVE (resource create-cve)

```bash
go run cmd/cli/secra.go resource create-cve \
  --source-id 7892cf6a-5b71-4496-a25a-8de9d58a8fa2 \
  --source-uid "CVE-2025-12345" \
  --title "Sample vulnerability" \
  --description "Sample vulnerability"
```

輸出範例：

```bash
CVE created: ID=<cve-uuid> SourceID=<source-uuid> SourceUID=CVE-2025-12345
```

## 5. 訂閱 Vendor (resource subscribe-vendor)

```bash
go run cmd/cli/secra.go resource subscribe-vendor \
  --user-id "43b5d7a3-958c-40c4-8237-c501a46c9ff5" \
  --vendor-id "3d9ed458-4df8-478e-81f3-5bb5e35b9192" \
  --severity "medium"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> Vendors=[<vendor-uuid>] Severity=medium
```

## 6. 訂閱 Product (resource subscribe-product)

```bash
go run cmd/cli/secra.go resource subscribe-product \
  --user-id "43b5d7a3-958c-40c4-8237-c501a46c9ff5" \
  --product-id "3f2ddd4e-8534-4af1-a959-4c54c5cdc226" \
  --severity "low"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> Products=[<product-uuid>] Severity=low
```

## 7. 訂閱 CVE Resource (resource subscribe-cve-resource)

```bash
go run cmd/cli/secra.go resource subscribe-cve-resource \
  --user-id "43b5d7a3-958c-40c4-8237-c501a46c9ff5" \
  --resource-id "7892cf6a-5b71-4496-a25a-8de9d58a8fa2" \
  --severity "high"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> CVEResources=[<resource-uuid>] Severity=high
