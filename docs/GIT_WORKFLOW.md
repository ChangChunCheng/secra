# Secra Git 工作流與開發規範

本文件定義了 `Secra` 專案的 Git 管理規則，所有開發者（包含 AI Agent）必須嚴格遵守。

## 1. 分支策略 (Branching Model)

專案採用簡化版的 Git Flow：

- **`main`**: 穩定版本分支。僅接受來自 `develop` 或 `hotfix/` 的合併。此分支的每個提交都應對應一個 Release Tag。
- **`develop`**: 開發整合分支。所有新功能與一般修復最終都會匯集於此。
- **`feature/ <name>`**: 新功能開發分支。從 `develop` 分開，完成後合併回 `develop`。
- **`fix/ <name>`**: 一般 Bug 修復分支。從 `develop` 分開，完成後合併回 `develop`。
- **`hotfix/ <version>`**: 緊急修復分支。直接從 `main` 分開，修復後同時合併回 `main` 與 `develop`。
- **`release/ <version>`**: 發佈準備分支。從 `develop` 分開，進行版本號更新與最終測試，完成後合併回 `main`。

## 2. 提交訊息規範 (Commit Message)

採用 **Conventional Commits** 格式：`<type>(<scope>): <description>`

- **`feat`**: 新功能 (Feature)
- **`fix`**: 修復 Bug (Bug Fix)
- **`docs`**: 僅文件變更 (Documentation)
- **`style`**: 程式碼格式調整（不影響邏輯，如空白、縮排）
- **`refactor`**: 程式碼重構（非修復也非新功能）
- **`perf`**: 效能優化 (Performance)
- **`test`**: 新增或修正測試 (Testing)
- **`chore`**: 建構流程、依賴工具變更（如 Makefile, Dockerfile 更新）

**範例**: `feat(cve): add cvss score column to list view`

## 3. 版本號與標籤 (Versioning & Tagging)

採用 **語義化版本 (Semantic Versioning 2.0.0)**：`vMAJOR.MINOR.PATCH`

- **MAJOR**: 當有不相容的 API 變更。
- **MINOR**: 當以回溯相容的方式新增功能。
- **PATCH**: 當以回溯相容的方式修復 Bug。
- **Pre-release**: 如 `-alpha.N`, `-beta.N`, `-rc.N`。

**標籤操作**:
- 標籤必須是 **Annotated Tag**：`git tag -a v0.1.0 -m "Release description"`。
- 版本號是系統 Build 標記的唯一來源（由 Makefile 自動抓取）。

## 4. Release 流程

1. 從 `develop` 建立 `release/vX.Y.Z` 分支。
2. 執行 `make build` 與 E2E 測試，確保版本資訊正確注入。
3. 更新 `CHANGELOG.md` (若有)。
4. 合併回 `main` 並打上標籤。
5. 將標籤合併回 `develop` 確保版本同步。

---
*遵循此規則，確保專案長期的穩定與可維護性。*
