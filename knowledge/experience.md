# 經驗

<!--
  這份文件記錄從開發過程中蒸餾出的教訓——不是 changelog，
  而是應該影響未來決策的模式。

  每個教訓記錄「理論」和「現實」之間的落差。
  保持簡短、可操作。詳細的事件記錄放在 knowledge/history/。
-->

## 教訓

### 沿用既有檔案而非依規格建立新檔案

- **理論說**：規格文件為每個功能規劃了獨立的檔案（例如 `api/api_keys.go`、`api/backup.go`、`api/mcp.go`）
- **實際發生**：專案已有一個統一的 `handlers.go` 處理所有 API handler，建立獨立檔案會打破既有的程式碼組織模式
- **解決方式**：將新功能的 handler 加入既有的 `handlers.go`，只在真正需要獨立模組時（如 `mcp/` 目錄）才建立新檔案
- **教訓**：規格是指引而非教條。實作時應優先遵循既有程式碼的組織慣例，而非機械地按規格建檔
- **來源**：commit f662a40

### 資料庫 schema 已預先建立的功能可以直接使用

- **理論說**：實作 API Key 功能需要先建立資料庫 migration
- **實際發生**：`001_init.up.sql` 早在 Phase 2 就已經包含 `api_keys` 表的完整 schema，中介軟體也已有 `APIKeyAuth` 的骨架
- **解決方式**：直接使用既有 schema，只需實作 model、repository、service 層
- **教訓**：動手前先檢查既有基礎設施。前期規劃可能已經鋪好了路，避免重複工作
- **來源**：backend/database/migrations/001_init.up.sql
