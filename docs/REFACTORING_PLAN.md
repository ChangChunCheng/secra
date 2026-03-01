# SECRA 專案重構與測試規劃

## 現況分析

### 當前專案結構
```
secra/
├── frontend/            # Next.js 前端 (完整的獨立應用)
├── cmd/                 # Go 後端入口
├── internal/            # Go 後端邏輯
├── api/                 # Protobuf API 定義 (前後端共用)
├── migrations/          # 資料庫遷移 (後端專屬)
├── scripts/             # 運維腳本 (後端專屬)
├── tests/e2e/           # E2E 測試 (已清空)
├── go.mod, go.sum       # Go 依賴 (後端)
├── Dockerfile           # 後端 Docker 配置
├── frontend/Dockerfile  # 前端 Docker 配置
└── docker-compose.yml   # 服務編排 (混合)
```

### 問題點
1. **結構混亂**: 前後端代碼在根目錄混雜
2. **測試缺失**: 
   - 後端只有 1 個單元測試
   - E2E 測試已刪除但目錄結構還在
   - 前端無測試
3. **文檔不清**: API 文檔、架構說明分散
4. **構建配置**: Dockerfile 和 docker-compose.yml 混在根目錄

---

## 重構方案

### 選項 A: Monorepo 結構 (推薦)
保持單一倉庫，但明確劃分前後端邊界。

```
secra/                          # 專案根目錄
├── backend/                    # 後端完整應用
│   ├── cmd/                    # 應用入口
│   │   ├── server/             # HTTP + gRPC 服務器
│   │   └── cli/                # CLI 工具
│   ├── internal/               # 私有業務邏輯
│   │   ├── api/web/            # REST API handlers
│   │   ├── service/            # 業務服務層
│   │   ├── repo/               # 資料庫層
│   │   ├── model/              # 數據模型
│   │   ├── auth/               # 認證
│   │   ├── config/             # 配置
│   │   └── ...
│   ├── migrations/             # 資料庫遷移
│   ├── scripts/                # 運維腳本
│   ├── tests/                  # 後端測試
│   │   ├── unit/               # 單元測試
│   │   ├── integration/        # 集成測試
│   │   └── fixtures/           # 測試數據
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   ├── Makefile
│   └── README.md
│
├── frontend/                   # 前端完整應用
│   ├── src/
│   ├── public/
│   ├── tests/                  # 前端測試
│   │   ├── unit/               # 單元測試 (Jest)
│   │   ├── integration/        # 組件測試 (React Testing Library)
│   │   └── e2e/                # E2E 測試 (Playwright)
│   ├── package.json
│   ├── Dockerfile
│   └── README.md
│
├── api/                        # 共用 API 定義
│   ├── v1/*.proto              # Protobuf 定義
│   ├── buf.yaml
│   └── buf.gen.yaml
│
├── tests/                      # 整合測試
│   └── e2e/                    # 全系統 E2E 測試
│       ├── specs/              # 測試規格
│       ├── fixtures/           # 測試數據
│       └── playwright.config.ts
│
├── docs/                       # 專案文檔
│   ├── architecture.md         # 架構說明
│   ├── api-reference.md        # API 參考
│   ├── deployment.md           # 部署指南
│   └── development.md          # 開發指南
│
├── deployments/                # 部署配置
│   ├── docker/
│   │   ├── backend.Dockerfile
│   │   ├── frontend.Dockerfile
│   │   └── docker-compose.yml
│   └── kubernetes/             # K8s 配置 (未來)
│
├── .github/                    # CI/CD 配置
│   └── workflows/
│       ├── backend-tests.yml
│       ├── frontend-tests.yml
│       └── e2e-tests.yml
│
├── .gitignore
├── README.md                   # 專案總覽
└── LICENSE
```

**優點**:
- 前後端完全獨立，各自有完整的目錄結構
- 清晰的測試層級結構
- 符合 Monorepo 最佳實踐
- 便於 CI/CD 管道分離

**缺點**:
- 需要較大規模的檔案移動
- 路徑引用需要更新

---

### 選項 B: 最小變動方案 (快速)
保持當前結構，僅增加測試和文檔。

```
secra/                          # 保持不變
├── tests/
│   ├── backend/
│   │   ├── unit/
│   │   └── integration/
│   ├── frontend/
│   │   ├── unit/
│   │   └── integration/
│   └── e2e/
│       └── specs/
└── docs/
    ├── architecture.md
    └── api-reference.md
```

**優點**:
- 改動最小，不破壞現有結構
- 快速實施

**缺點**:
- 結構仍然不夠清晰
- 前後端邊界模糊

---

## 測試策略

### 後端測試架構

#### 1. 單元測試 (Unit Tests)
**目標**: 測試單個函數/方法的邏輯

**工具**: Go 內建 `testing` + `testify/assert`

**範例結構**:
```
backend/tests/unit/
├── service/
│   ├── user_service_test.go
│   ├── cve_service_test.go
│   └── subscription_service_test.go
├── repo/
│   ├── user_repo_test.go
│   └── cve_repo_test.go
├── auth/
│   └── hash_test.go
└── util/
    └── uuid_test.go
```

**重點測試項目**:
- ✅ `auth.HashPassword()` / `CheckPasswordHash()`
- ✅ `service.UserService.Register()`
- ✅ `service.UserService.Login()`
- ✅ `service.SubscriptionService.Create()`
- ✅ `repo.UserRepository.CreateUser()`
- ✅ Business logic: 訂閱匹配邏輯
- ✅ Notification batching logic

**範例**:
```go
func TestUserService_Register(t *testing.T) {
    mockRepo := &MockUserRepository{}
    userSvc := service.NewUserService(mockRepo)
    
    user, err := userSvc.Register(context.Background(), "test", "test@example.com", "password123", "password123")
    
    assert.NoError(t, err)
    assert.Equal(t, "test", user.Username)
    assert.NotEmpty(t, user.PasswordHash)
    mockRepo.AssertCalled(t, "CreateUser")
}
```

#### 2. 集成測試 (Integration Tests)
**目標**: 測試多個組件協作，包含真實資料庫

**工具**: `testcontainers-go` (Docker 容器測試)

**範例結構**:
```
backend/tests/integration/
├── api/
│   ├── auth_test.go            # POST /api/v1/auth/login
│   ├── cve_test.go             # GET /api/v1/cves
│   └── subscription_test.go    # POST /api/v1/subscriptions
├── database/
│   └── migration_test.go       # 測試遷移腳本
└── fixtures/
    ├── users.sql
    └── cves.sql
```

**重點測試項目**:
- ✅ API 端點完整流程 (request → response)
- ✅ 資料庫 CRUD 操作
- ✅ JWT 認證流程
- ✅ CORS 配置驗證

**範例**:
```go
func TestAPI_Login_Integration(t *testing.T) {
    // Start test database container
    ctx := context.Background()
    pgContainer, _ := postgres.RunContainer(ctx)
    defer pgContainer.Terminate(ctx)
    
    // Setup test server
    db := setupTestDB(pgContainer.ConnectionString())
    server := setupTestServer(db)
    
    // Test login
    resp := httptest.NewRequest("POST", "/api/v1/auth/login", loginPayload)
    w := httptest.NewRecorder()
    server.ServeHTTP(w, resp)
    
    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Header().Get("Set-Cookie"), "session_token")
}
```

---

### 前端測試架構

#### 1. 單元測試 (Unit Tests)
**目標**: 測試純函數、utilities、Redux slices

**工具**: Jest + React Testing Library

**範例結構**:
```
frontend/tests/unit/
├── lib/
│   ├── features/
│   │   ├── authSlice.test.ts
│   │   └── apiSlice.test.ts
│   └── utils/
│       └── formatDate.test.ts
└── components/
    ├── Pagination.test.tsx
    └── Navbar.test.tsx
```

**重點測試項目**:
- ✅ Redux reducers (authSlice)
- ✅ API query transformations
- ✅ 純函數組件 (Pagination, ViewToggle)
- ✅ Utility functions

**範例**:
```typescript
// authSlice.test.ts
describe('authSlice', () => {
  it('should set user on login', () => {
    const state = authReducer(initialState, setUser({ username: 'test' }));
    expect(state.user).toEqual({ username: 'test' });
    expect(state.isAuthenticated).toBe(true);
  });
});

// Pagination.test.tsx
describe('Pagination', () => {
  it('renders page numbers correctly', () => {
    render(<Pagination currentPage={1} totalPages={5} onPageChange={jest.fn()} />);
    expect(screen.getByText('1')).toBeInTheDocument();
    expect(screen.getByText('5')).toBeInTheDocument();
  });
});
```

#### 2. 組件測試 (Component Tests)
**目標**: 測試 React 組件的交互行為

**工具**: React Testing Library + MSW (Mock Service Worker)

**範例結構**:
```
frontend/tests/integration/
├── pages/
│   ├── LoginPage.test.tsx
│   ├── Dashboard.test.tsx
│   └── ProfilePage.test.tsx
└── mocks/
    ├── handlers.ts             # MSW API mocks
    └── server.ts
```

**重點測試項目**:
- ✅ 登入流程 (表單填寫 → API 調用 → 重定向)
- ✅ 訂閱管理 (新增/刪除/更新)
- ✅ 分頁和過濾功能
- ✅ 錯誤處理 (401, 404, 500)

**範例**:
```typescript
// LoginPage.test.tsx
describe('LoginPage', () => {
  it('successful login redirects to dashboard', async () => {
    const user = userEvent.setup();
    render(<LoginPage />);
    
    await user.type(screen.getByLabelText(/username/i), 'admin');
    await user.type(screen.getByLabelText(/password/i), 'admin');
    await user.click(screen.getByRole('button', { name: /login/i }));
    
    await waitFor(() => {
      expect(mockRouter.push).toHaveBeenCalledWith('/');
    });
  });
});
```

#### 3. E2E 測試 (End-to-End Tests)
**目標**: 測試完整用戶流程

**工具**: Playwright

**範例結構**:
```
frontend/tests/e2e/
├── specs/
│   ├── auth.spec.ts            # 註冊、登入、登出
│   ├── cve-browsing.spec.ts    # 瀏覽、搜尋、過濾
│   ├── subscription.spec.ts    # 訂閱管理
│   └── admin.spec.ts           # 管理員功能
├── fixtures/
│   └── test-data.ts
└── playwright.config.ts
```

**重點測試項目**:
- ✅ 用戶註冊 → 登入 → 訂閱產品 → 查看儀表板
- ✅ 搜尋和過濾 CVE
- ✅ 分頁導航
- ✅ 管理員用戶管理
- ✅ Profile 更新

**範例**:
```typescript
// subscription.spec.ts
test('user can subscribe to a product', async ({ page }) => {
  await loginAsUser(page, 'testuser', 'password');
  
  await page.goto('/products');
  await page.click('text=Linux Kernel');
  await page.click('button:has-text("Subscribe")');
  
  await page.goto('/my/dashboard');
  await expect(page.locator('text=Linux Kernel')).toBeVisible();
});
```

---

### 整合 E2E 測試
**目標**: 測試前後端完整集成

**工具**: Playwright + Docker Compose

**範例結構**:
```
tests/e2e/
├── specs/
│   ├── full-workflow.spec.ts   # 完整用戶旅程
│   ├── api-integration.spec.ts # API 正確性
│   └── notification.spec.ts    # 通知系統
├── docker-compose.test.yml     # 測試環境配置
└── setup.ts                    # 測試前置作業
```

**測試流程**:
1. 啟動測試環境 (`docker-compose -f docker-compose.test.yml up`)
2. 執行資料庫遷移和初始化
3. 填充測試數據
4. 執行 Playwright 測試
5. 清理環境

---

## 實施計劃

### Phase 1: 目錄重構 (2-3 天)
**任務**:
1. 創建新的目錄結構 (`backend/`, `deployments/`, `docs/`)
2. 移動檔案到對應位置
3. 更新所有 import 路徑
4. 更新 Dockerfile 和 docker-compose.yml 路徑
5. 更新 Makefile
6. 更新 README.md

**驗證**:
- ✅ `make build` 成功
- ✅ `make docker-up` 成功啟動
- ✅ 前端可訪問 http://localhost
- ✅ API 正常回應

### Phase 2: 後端測試 (3-4 天)
**任務**:
1. 設置測試框架 (`testify`, `testcontainers-go`)
2. 編寫單元測試 (優先: auth, user service, subscription service)
3. 編寫集成測試 (優先: auth API, subscription API)
4. 設置 CI/CD 測試管道

**目標覆蓋率**: >70%

### Phase 3: 前端測試 (3-4 天)
**任務**:
1. 設置 Jest + React Testing Library
2. 編寫組件單元測試
3. 設置 MSW mocks
4. 編寫頁面集成測試
5. 設置 Playwright E2E 測試

**目標覆蓋率**: >60%

### Phase 4: 整合 E2E 測試 (2 天)
**任務**:
1. 創建測試環境配置
2. 編寫完整用戶旅程測試
3. 自動化測試執行

### Phase 5: 文檔與優化 (1-2 天)
**任務**:
1. 編寫架構文檔
2. 編寫 API 文檔 (Swagger/OpenAPI)
3. 更新開發指南
4. Code review 和重構

---

## 配置文件範例

###  Backend Makefile
```makefile
.PHONY: test test-unit test-integration test-coverage

test: test-unit test-integration

test-unit:
	@echo "🧪 Running unit tests..."
	go test -v -short ./tests/unit/...

test-integration:
	@echo "🧪 Running integration tests..."
	go test -v -run Integration ./tests/integration/...

test-coverage:
	@echo "📊 Generating coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
```

### Frontend package.json scripts
```json
{
  "scripts": {
    "test": "jest",
    "test:watch": "jest --watch",
    "test:coverage": "jest --coverage",
    "test:e2e": "playwright test",
    "test:e2e:ui": "playwright test --ui"
  }
}
```

### GitHub Actions (CI)
```yaml
# .github/workflows/backend-tests.yml
name: Backend Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:17-alpine
        env:
          POSTGRES_PASSWORD: postgres
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: cd backend && go test -v ./...
```

---

## 建議優先順序

### 立即執行 (高優先級)
1. ✅ **選擇重構方案**: 建議採用 **選項 A (Monorepo 結構)**
2. ✅ **後端單元測試**: 先補齊核心業務邏輯測試
3. ✅ **API 集成測試**: 驗證所有 REST API 端點

### 短期執行 (中優先級)
4. ✅ **前端組件測試**: 測試關鍵頁面和組件
5. ✅ **目錄重構**: 移動到新結構

### 長期執行 (低優先級)
6. ✅ **E2E 測試**: 完整用戶旅程測試
7. ✅ **性能測試**: 負載測試和壓力測試
8. ✅ **安全測試**: 漏洞掃描和滲透測試

---

## 總結

**推薦方案**: 選項 A (Monorepo 結構) + 完整測試架構

**預期成果**:
- 📁 清晰的前後端分離目錄結構
- ✅ >70% 後端測試覆蓋率
- ✅ >60% 前端測試覆蓋率
- 🤖 自動化 CI/CD 測試管道
- 📚 完整的專案文檔

**時間估計**: 10-15 天 (取決於團隊規模和測試深度)
