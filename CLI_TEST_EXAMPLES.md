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

## 3. 取得個人資料 (user get-profile)

```bash
go run cmd/cli/secra.go user get-profile \
  --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ7XCJpZFwiOlwiNTc4N2U3MTItNmUzOS00OWIyLTk5NDctNDNiODJiYzg2MGUxXCIsXCJ1c2VybmFtZVwiOlwiYWxpY2VcIixcInJvbGVcIjpcInVzZXJcIn0iLCJleHAiOjE3NTI2NTYxODksImlhdCI6MTc1MjU2OTc4OX0.lo4VNJ8RqKc6-H49KbgBnnbXFj4UckNggT2MGCeq16U"
```

輸出範例：

```bash
ID: 123e4567-e89b-12d3-a456-426614174000
Username: alice
Email: alice@example.com
```

## 4. 更新使用者個人資料 (user update-profile)

```bash
go run cmd/cli/secra.go user update-profile \
  --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ7XCJpZFwiOlwiNTc4N2U3MTItNmUzOS00OWIyLTk5NDctNDNiODJiYzg2MGUxXCIsXCJ1c2VybmFtZVwiOlwiYWxpY2VcIixcInJvbGVcIjpcInVzZXJcIn0iLCJleHAiOjE3NTI2NTYxODksImlhdCI6MTc1MjU2OTc4OX0.lo4VNJ8RqKc6-H49KbgBnnbXFj4UckNggT2MGCeq16U" \
  --email "new@example.com" 
```

輸出範例：

```bash
Updated Profile:
ID: 123e4567-e89b-12d3-a456-426614174000
Username: alice
Email: new@example.com
```

## 4. 新增 CVE 資源 (resource create-cve-source)

```bash
go run cmd/cli/secra.go resource create-cve-source \
  --name "ttt" \
  --type "VendorX Source" \
  --url "https://vendorx.com/feed" \
  --description "https://vendorx.com/feed/desc"
```

輸出範例：

```bash
Created CVE resource: ID=e80e0407-a660-4e19-a83a-bb2070707ccc Name=ttt URL=https://vendorx.com/feed
```

## 5. 新增 Vendor (resource create-vendor)

```bash
go run cmd/cli/secra.go resource create-vendor \
  --name "VendorX"
```

輸出範例：

```bash
Created Vendor: ID=<vendor-uuid> Name=VendorX
```

輸出範例：

```bash
CVE created: ID=<cve-uuid> SourceID=<source-uuid> SourceUID=CVE-2025-12345
```

## 8. 取得 Vendor (resource get-vendor)

```bash
go run cmd/cli/secra.go resource get-vendor \
  --id "dc791f4c-f0f3-4ff1-89e3-a54dc592446a"
```

輸出範例：

```bash
ID=<vendor-uuid> Name=VendorX
```

## 9. 列出 Vendors (resource list-vendor)

```bash
go run cmd/cli/secra.go resource list-vendor \
  --limit 5 \
  --offset 0
```

輸出範例：

```bash
ID=uuid1 Name=VendorA
ID=uuid2 Name=VendorB
```

## 10. 更新 Vendor (resource update-vendor)

```bash
go run cmd/cli/secra.go resource update-vendor \
  --id "<vendor-uuid>" \
  --name "NewVendorName"
```

輸出範例：

```bash
Updated Vendor: ID=<vendor-uuid> Name=NewVendorName
```

## 11. 刪除 Vendor (resource delete-vendor)

```bash
go run cmd/cli/secra.go resource delete-vendor \
  --id "dc791f4c-f0f3-4ff1-89e3-a54dc592446a"
```

輸出範例：

```bash
Deleted Vendor: ID=<vendor-uuid>
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

## NVD 匯入範例

## NVD v1 最近資料
```bash
go run cmd/cli/secra.go import nvd v1 --recent=true
```

## NVD v1 指定時間區間

```bash
go run cmd/cli/secra.go import nvd v1 --start=2025-01-01 --end=2025-01-31
```

## NVD v2 匯入

```bash
go run cmd/cli/secra.go import nvd v2 --start=2025-01-01 --apikey=$$NVD_API_KEY
```
