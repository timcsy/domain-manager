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

### 操作 CRD 需要 Dynamic Client

- **理論說**：client-go 的 Clientset 可以操作所有 K8s 資源
- **實際發生**：cert-manager 的 ClusterIssuer 是 CRD，標準 Clientset 無法操作。需要引入 `k8s.io/client-go/dynamic` 搭配 `unstructured.Unstructured` 才能 CRUD CRD 資源
- **解決方式**：在 `k8s/client.go` 初始化時同時建立 `DynamicClient`，供 `certmanager.go` 使用
- **教訓**：操作第三方 CRD（cert-manager、Traefik IngressRoute 等）時，必須使用 dynamic client，不能假設 Clientset 就夠用
- **來源**：commit 0dcdff7, backend/src/k8s/certmanager.go

### 敏感資料需要雙重儲存策略

- **理論說**：Cloudflare API Token 存在系統設定（SQLite）就夠了
- **實際發生**：cert-manager 需要從 K8s Secret 讀取 token，不會去查應用程式的 SQLite。因此 token 必須同時存在 system_settings（應用層讀取）和 K8s Secret（cert-manager 讀取）
- **解決方式**：`SaveToken` 同時寫入兩處，`RemoveToken` 同時清除兩處
- **教訓**：當資料需要被多個系統消費時，考慮每個消費者的存取方式，可能需要同步到不同儲存層
- **來源**：commit 0dcdff7, backend/src/services/cloudflare_service.go

### Alpine.js inline x-data 不支援 async method shorthand

- **理論說**：JavaScript object literal 支援 `async foo() {}` method shorthand，所以 Alpine.js 的 `x-data` 也應該支援
- **實際發生**：Alpine.js 3.x 用 `new Function()` 解析 inline x-data，不支援 async shorthand 語法。頁面完全壞掉，所有變數都 undefined
- **解決方式**：改用 `foo: async function() {}` function expression 語法
- **教訓**：框架的模板語法不等於原生 JavaScript。在框架的 eval 環境中，語法支援可能有限制，遇到 parse error 優先懷疑語法相容性
- **來源**：commit 025c4e7, frontend/src/pages/settings.html

### 用功能性端點驗證第三方 API token 更穩健

- **理論說**：Cloudflare 提供 `/user/tokens/verify` 端點驗證 API Token
- **實際發生**：Cloudflare 新版 `cfat_` 格式的 token（從 Edit zone DNS template 建立）在 `/user/tokens/verify` 回傳 401 Invalid。舊版文件的驗證方式不適用於新格式
- **解決方式**：改用 `/zones?per_page=1` 驗證——直接呼叫 token 實際需要的功能端點，同時確認權限正確
- **教訓**：驗證第三方 API token 時，優先用功能性端點（實際會用到的 API）而非專用驗證端點。格式可能變化，但功能端點的行為更穩定
- **來源**：commit 8ec0fec, backend/src/services/cloudflare_service.go
