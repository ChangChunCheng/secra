# secra CLI 操作說明與範例

以下示範如何使用各子指令進行操作，需以 `go run cmd/cli/secra.go` 作為前綴執行，假設使用 gvm 安裝的 Go 執行環境已正確設定。

---

## 1. 使用者 (User) 子指令

### 1.1 註冊 (register)

```bash
go run cmd/cli/secra.go user register \
  --username alice \
  --email alice@example.com \
  --password "securepassword"
```

**輸出範例：**

```bash
Registered user: alice (id=<user-uuid>)
```

### 1.2 登入 (login)

```bash
go run cmd/cli/secra.go user login \
  --username alice \
  --password "securepassword"
```

**輸出範例：**

```bash
75f05a92-db6e-4f5d-91b9-a25333495f89
```

### 1.3 取得個人資料 (get_profile)

```bash
go run cmd/cli/secra.go user get_profile \
  --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**輸出範例：**

```bash
ID: 123e4567-e89b-12d3-a456-426614174000
Username: alice
Email: alice@example.com
DisplayName: Alice Chen
```

### 1.4 更新個人資料 (update_profile)

```bash
go run cmd/cli/secra.go user update_profile \
  --token "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  --email new@example.com \
  --display-name "Alice C."
```

**輸出範例：**

```bash
Updated profile: alice
```

---

## 2. 匯入 (Import) 子指令

### 2.1 NVD 匯入 (nvd)

```bash
go run cmd/cli/secra.go import nvd \
  --source v2 \
  --since 2025-01-01
```

**輸出範例：**

```bash
Imported 120 CVEs from NVD v2 since 2025-01-01
```

---

## 3. 資料庫遷移 (Migrate) 子指令

### 3.1 顯示遷移狀態 (status)

```bash
go run cmd/cli/secra.go migrate status
```

**輸出範例：**

```bash
Pending migrations:
  - 20230101_add_new_table.up.sql
Applied migrations:
  - 20230101_initial_schema.up.sql
```

### 3.2 執行遷移 (up)

```bash
go run cmd/cli/secra.go migrate up
```

**輸出範例：**

```bash
Applying migration: 20230101_add_new_table.up.sql
Migration completed.
```

---

## 4. 資源 (Resource) 子指令

### 4.1 CVE (cve)

- **列出 (list)**

  ```bash
  go run cmd/cli/secra.go resource cve list
  ```

- **取得 (get)**

  ```bash
  go run cmd/cli/secra.go resource cve get \
    --id CVE-2025-1234
  ```

- **新增 (create)**

  ```bash
  go run cmd/cli/secra.go resource cve create \
    --id CVE-2025-9999 \
    --description "範例 CVE"
  ```

- **更新 (update)**

  ```bash
  go run cmd/cli/secra.go resource cve update \
    --id CVE-2025-9999 \
    --description "更新後描述"
  ```

- **刪除 (delete)**

  ```bash
  go run cmd/cli/secra.go resource cve delete \
    --id CVE-2025-9999
  ```

### 4.2 CVE Source (cvesource)

- **列出 (list)**

  ```bash
  go run cmd/cli/secra.go resource cvesource list
  ```

- **取得 (get)**

  ```bash
  go run cmd/cli/secra.go resource cvesource get \
    --id nvd-v2
  ```

- **新增 (create)**

  ```bash
  go run cmd/cli/secra.go resource cvesource create \
    --id nvd-v3 \
    --url https://example.com/nvd.json
  ```

- **更新 (update)**

  ```bash
  go run cmd/cli/secra.go resource cvesource update \
    --id nvd-v3 \
    --url https://example.com/nvd-updated.json
  ```

- **刪除 (delete)**

  ```bash
  go run cmd/cli/secra.go resource cvesource delete \
    --id nvd-v3
  ```

### 4.3 產品 (product)

- **列出 (list)**

  ```bash
  go run cmd/cli/secra.go resource product list
  ```

- **取得 (get)**

  ```bash
  go run cmd/cli/secra.go resource product get \
    --id prod-001
  ```

- **新增 (create)**

  ```bash
  go run cmd/cli/secra.go resource product create \
    --id prod-002 \
    --name "新產品"
  ```

- **更新 (update)**

  ```bash
  go run cmd/cli/secra.go resource product update \
    --id prod-002 \
    --name "更新後產品"
  ```

- **刪除 (delete)**

  ```bash
  go run cmd/cli/secra.go resource product delete \
    --id prod-002
  ```

### 4.4 訂閱 (subscription)

- **列出 (list)**

  ```bash
  go run cmd/cli/secra.go resource subscription list
  ```

- **建立 (create)**

  ```bash
  go run cmd/cli/secra.go resource subscription create \
    --source-id nvd-v2 \
    --target-id prod-001 \
    --type product
  ```

- **刪除 (delete)**

  ```bash
  go run cmd/cli/secra.go resource subscription delete \
    --id sub-123
  ```
