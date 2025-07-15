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
75f05a92-db6e-4f5d-91b9-a25333495f89
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
  --source-id bff832d2-002e-41b9-988f-90f930277a58 \
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
  --user-id "5787e712-6e39-49b2-9947-43b82bc860e1" \
  --vendor-id "0e029aaa-d339-413a-af9e-f1bb26c9a1f8" \
  --severity "medium"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> Vendors=[<vendor-uuid>] Severity=medium
```

## 6. 訂閱 Product (resource subscribe-product)

```bash
go run cmd/cli/secra.go resource subscribe-product \
  --user-id "5787e712-6e39-49b2-9947-43b82bc860e1" \
  --product-id "63637e30-0c91-419c-bd46-ad68650dacf9" \
  --severity "low"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> Products=[<product-uuid>] Severity=low
```

## 7. 訂閱 CVE Resource (resource subscribe-cve-resource)

```bash
go run cmd/cli/secra.go resource subscribe-cve-resource \
  --user-id "5787e712-6e39-49b2-9947-43b82bc860e1" \
  --resource-id "bff832d2-002e-41b9-988f-90f930277a58" \
  --severity "high"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> CVEResources=[<resource-uuid>] Severity=high
