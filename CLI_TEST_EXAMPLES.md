
# CLI 子功能測試指令範例

以下示範如何使用各子指令進行操作，需以 [`go run cmd/cli/secra.go`](cmd/cli/secra.go:1) 作為前綴執行，假設使用 gvm 安裝的 Go 執行環境已正確設定。

---

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
  --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
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
  --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  --email "new@example.com"
```

輸出範例：

```bash
Updated Profile:
ID: 123e4567-e89b-12d3-a456-426614174000
Username: alice
Email: new@example.com
```

---

## 5. 管理 CVE Source (resource cve-source)

### 5.1 新增 CVE Source

```bash
go run cmd/cli/secra.go resource cve-source create \
  --name "ttt" \
  --type "VendorX Source" \
  --url "https://vendorx.com/feed" \
  --description "https://vendorx.com/feed/desc"
```

輸出範例：

```bash
Created CVE source: ID=e80e0407-a660-4e19-a83a-bb2070707ccc Name=ttt URL=https://vendorx.com/feed
```

### 5.2 取得 CVE Source (resource cve-source get)

```bash
go run cmd/cli/secra.go resource cve-source get \
  --id "e80e0407-a660-4e19-a83a-bb2070707ccc"
```

輸出範例：

```bash
ID=e80e0407-a660-4e19-a83a-bb2070707ccc Name=ttt Type=VendorX Source URL=https://vendorx.com/feed Description=https://vendorx.com/feed/desc
```

### 5.3 列出 CVE Sources (resource cve-source list)

```bash
go run cmd/cli/secra.go resource cve-source list \
  --limit 5 \
  --offset 0
```

輸出範例：

```bash
ID=uuid1 Name=SourceA
ID=uuid2 Name=SourceB
```

### 5.4 更新 CVE Source (resource cve-source update)

```bash
go run cmd/cli/secra.go resource cve-source update \
  --id "e80e0407-a660-4e19-a83a-bb2070707ccc" \
  --name "newName" \
  --url "https://new.example.com"
```

輸出範例：

```bash
Updated CVE source: ID=e80e0407-a660-4e19-a83a-bb2070707ccc Name=newName URL=https://new.example.com
```

### 5.5 刪除 CVE Source (resource cve-source delete)

```bash
go run cmd/cli/secra.go resource cve-source delete \
  --id "e80e0407-a660-4e19-a83a-bb2070707ccc"
```

輸出範例：

```bash
Deleted CVE source: ID=e80e0407-a660-4e19-a83a-bb2070707ccc
```

---

## 6. 管理 Vendor (resource vendor)

### 6.1 新增 Vendor

```bash
go run cmd/cli/secra.go resource vendor create \
  --name "VendorX"
```

輸出範例：

```bash
Created Vendor: ID=<vendor-uuid> Name=VendorX
```

### 6.2 取得 Vendor (resource vendor get)

```bash
go run cmd/cli/secra.go resource vendor get \
  --id "dc791f4c-f0f3-4ff1-89e3-a54dc592446a"
```

輸出範例：

```bash
ID=<vendor-uuid> Name=VendorX
```

### 6.3 列出 Vendors (resource vendor list)

```bash
go run cmd/cli/secra.go resource vendor list \
  --limit 5 \
  --offset 0
```

輸出範例：

```bash
ID=uuid1 Name=VendorA
ID=uuid2 Name=VendorB
```

### 6.4 更新 Vendor (resource vendor update)

```bash
go run cmd/cli/secra.go resource vendor update \
  --id "<vendor-uuid>" \
  --name "NewVendorName"
```

輸出範例：

```bash
Updated Vendor: ID=<vendor-uuid> Name=NewVendorName
```

### 6.5 刪除 Vendor (resource vendor delete)

```bash
go run cmd/cli/secra.go resource vendor delete \
  --id "dc791f4c-f0f3-4ff1-89e3-a54dc592446a"
```

輸出範例：

```bash
Deleted Vendor: ID=<vendor-uuid>
```

---

## 7. 管理 Product (resource product)

### 7.1 新增 Product

```bash
go run cmd/cli/secra.go resource product create \
  --name "ProductX"
```

輸出範例：

```bash
Created Product: ID=<product-uuid> Name=ProductX
```

### 7.2 取得 Product (resource product get)

```bash
go run cmd/cli/secra.go resource product get \
  --id "<product-uuid>"
```

輸出範例：

```bash
ID=<product-uuid> Name=ProductX
```

### 7.3 列出 Products (resource product list)

```bash
go run cmd/cli/secra.go resource product list \
  --limit 5 \
  --offset 0
```

輸出範例：

```bash
ID=uuid1 Name=ProductA
ID=uuid2 Name=ProductB
```

### 7.4 更新 Product (resource product update)

```bash
go run cmd/cli/secra.go resource product update \
  --id "<product-uuid>" \
  --name "NewProductName"
```

輸出範例：

```bash
Updated Product: ID=<product-uuid> Name=NewProductName
```

### 7.5 刪除 Product (resource product delete)

```bash
go run cmd/cli/secra.go resource product delete \
  --id "<product-uuid>"
```

輸出範例：

```bash
Deleted Product: ID=<product-uuid>
```

---

## 8. 訂閱功能 (resource subscribe)

以下示例使用 `resource subscribe` 子模組，可以對 Vendor、Product、CVE Source 進行訂閱。

### 8.1 訂閱 Vendor

```bash
go run cmd/cli/secra.go resource subscribe vendor \
  --user-id "5787e712-6e39-49b2-9947-43b82bc860e1" \
  --vendor-id "0e029aaa-d339-413a-af9e-f1bb26c9a1f8" \
  --severity "medium"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> Vendors=[<vendor-uuid>] Severity=medium
```

### 8.2 訂閱 Product

```bash
go run cmd/cli/secra.go resource subscribe product \
  --user-id "5787e712-6e39-49b2-9947-43b82bc860e1" \
  --product-id "63637e30-0c91-419c-bd46-ad68650dacf9" \
  --severity "low"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> Products=[<product-uuid>] Severity=low
```

### 8.3 訂閱 CVE Source

```bash
go run cmd/cli/secra.go resource subscribe cve-source \
  --user-id "5787e712-6e39-49b2-9947-43b82bc860e1" \
  --resource-id "bff832d2-002e-41b9-988f-90f930277a58" \
  --severity "high"
```

輸出範例：

```bash
Subscription created: User=<user-uuid> CVEResources=[<resource-uuid>] Severity=high
```

---

## 9. NVD 匯入範例

### 9.1 NVD v1 最近資料

```bash
go run cmd/cli/secra.go import nvd v1 --recent=true
```

### 9.2 NVD v1 指定時間區間

```bash
go run cmd/cli/secra.go import nvd v1 --start=2025-01-01 --end=2025-01-31
```

### 9.3 NVD v2 匯入

```bash
go run cmd/cli/secra.go import nvd v2 --start=2025-01-01 --apikey=$$NVD_API_KEY
```
