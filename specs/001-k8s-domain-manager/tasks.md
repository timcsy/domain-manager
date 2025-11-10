# 實作任務清單: Kubernetes 域名管理器

**功能分支**: `001-k8s-domain-manager`
**建立日期**: 2025-11-07
**更新日期**: 2025-11-10
**狀態**: In Progress - Phase 4.5
**版本**: 1.0

---

## 任務統計

- **總任務數**: 140 個
- **已完成**: 76 個 (54%)
- **進行中**: Phase 4.5 (CI/CD 與容器化)
- **Phase 1 (設定)**: 5 個任務 ✅ 完成
- **Phase 2 (基礎設施)**: 10 個任務 ✅ 完成
- **Phase 3 (US1)**: 26 個任務 ✅ 完成
- **Phase 4 (US2)**: 34 個任務 ⚡ 進行中 (27/34 完成)
- **Phase 4.5 (CI/CD)**: 8 個任務 ✅ 完成
- **Phase 5 (US3)**: 18 個任務
- **Phase 6 (US4)**: 31 個任務
- **Phase 7 (收尾)**: 6 個任務
- **可平行執行任務**: 90 個 (標記 [P])

## 最新進展 (2025-11-10)

### Phase 4.5 - CI/CD 與容器化 🚀

完成完整的 CI/CD 基礎設施建置，為 K8s 部署做好準備：

✅ **已完成**:
- 多階段 Dockerfile (Frontend + Backend，優化建置大小)
- Docker Compose 配置 (本地測試環境，Mock 模式支援)
- GitHub Actions CI workflow (自動測試、Lint、Docker 建置)
- GitHub Actions CD workflow (自動部署到 Staging 和 Production)
- Docker 清理腳本 (一鍵恢復系統)
- DOCKER.md 使用指南 (完整的 Docker 操作文件)
- .dockerignore (優化建置速度)

🧪 **測試結果**:
- Docker 建置測試: ✅ 成功 (~70秒)
- 容器運行測試: ✅ 成功
- API 測試: ✅ 登入與基本功能正常
- 清理測試: ✅ 完全恢復，無殘留資源

### Phase 4 - 核心功能優先策略

採用**選項 B - 核心功能優先**策略，實作基本的域名和憑證管理功能：

✅ **已完成**:
- K8s 整合層 (Ingress, Secret, Service, Health managers)
- Certificate model 和 repository
- Certificate Service (上傳、列表、查詢、刪除)
- Domain Service K8s 整合 (建立/更新/刪除 Ingress)
- Certificate API endpoints (4個)
- Service discovery API (2個)
- 所有核心 Domain API endpoints (5個)
- Diagnostic Service (系統健康檢查、日誌查詢、系統指標)
- Diagnostic API endpoints (健康檢查、日誌查詢)
- Certificate 前端頁面 (列表、上傳、刪除)
- Domain 前端頁面 (列表、新增表單、詳情頁面)
- UI 優化：Dashboard SSL 憑證數量動態顯示、側邊欄佈局修正

⏭️ **延後實作** (進階功能):
- Let's Encrypt 自動申請憑證 (T048-T050)
- 憑證自動續約排程器 (T056-T057)
- 憑證加密儲存 (T052)
- 到期憑證監控 (T069)

🎯 **測試結果**:
- 應用程式在 Mock 模式下成功運行 ✓
- Service discovery API 回傳模擬資料 ✓
- Certificate API 正常運作 ✓
- 資料庫 migrations 成功 ✓
- Dashboard 顯示實際 SSL 憑證數量 ✓
- 側邊欄導航佈局正常 ✓
- 憑證上傳與 K8s Secret 建立成功 ✓
- Domain 狀態 API 回傳真實憑證狀態 ✓

---

## Phase 1: 設定(專案初始化)

**目的**: 建立專案基礎結構和工具鏈

- [X] T001 建立後端專案目錄結構在 backend/src/{models,services,api,mcp,repositories,k8s,middleware,db}/
- [X] T002 [P] 初始化 go.mod 並新增依賴在 backend/go.mod (chi v5.0.0+, client-go v0.29.0+, lego v4.15.0+, modernc.org/sqlite v1.28.0+)
- [X] T003 [P] 建立前端專案目錄結構在 frontend/src/{components,pages,services,styles}/
- [X] T004 [P] 設定 TailwindCSS 和 HTMX 在 frontend/ (package.json, tailwind.config.js)
- [X] T005 [P] 建立 Helm Chart 目錄結構在 helm/domain-manager/{templates,charts}/ 和基礎檔案 (Chart.yaml, values.yaml)

---

## Phase 2: 基礎設施(阻塞性先決條件)

**目的**: 實作所有使用者故事依賴的核心基礎設施

⚠️ **關鍵**: 在此階段完成前,所有使用者故事無法開始

- [X] T006 建立 SQLite schema 在 backend/database/migrations/001_init.up.sql (包含 6 個表: domains, certificates, diagnostic_logs, admin_accounts, api_keys, system_settings)
- [X] T007 [P] 建立資料庫回滾腳本在 backend/database/migrations/001_init.down.sql
- [X] T008 [P] 實作資料庫連線管理在 backend/src/db/connection.go (SQLite 連接池、migration 執行)
- [X] T009 [P] 實作 K8s 客戶端初始化在 backend/src/k8s/client.go (in-cluster 和 kubeconfig 支援)
- [X] T010 [P] 實作認證中介軟體在 backend/src/middleware/auth.go (API Key 驗證、session 管理)
- [X] T011 [P] 實作日誌中介軟體在 backend/src/middleware/logging.go (請求/回應日誌)
- [X] T012 [P] 實作錯誤處理中介軟體在 backend/src/middleware/error.go (統一錯誤格式)
- [X] T013 實作基礎 API 路由結構在 backend/src/api/routes.go (Chi router 設定、中介軟體鏈)
- [X] T014 [P] 實作通用回應格式在 backend/src/api/response.go (Success, Error, Pagination helpers)
- [X] T015 實作主應用程式入口在 backend/src/main.go (服務啟動、graceful shutdown)

**檢查點**: 基礎設施就緒 - 使用者故事實作現在可以平行開始

---

## Phase 3: 使用者故事 1 - 透過 Helm 快速部署 🎯 MVP

**目標**: 使用者可以透過 Helm 在 K8s 上部署系統並存取 Web 介面

**獨立測試**: 在全新 K8s 叢集執行 `helm install domain-manager ./helm/domain-manager`,存取介面看到歡迎畫面

### 實作任務

#### 資料模型與 Repository (US1)

- [X] T016 [P] [US1] 建立 Domain 模型在 backend/src/models/domain.go (struct 定義、JSON tags、validation)
- [X] T017 [P] [US1] 建立 AdminAccount 模型在 backend/src/models/admin_account.go (struct 定義、密碼加密)
- [X] T018 [P] [US1] 建立 SystemSettings 模型在 backend/src/models/system_settings.go
- [X] T019 [US1] 實作 Domain repository 在 backend/src/repositories/domain_repo.go (Create, GetByID, GetByName, List, Update, Delete, SoftDelete, Count)
- [X] T020 [P] [US1] 實作 AdminAccount repository 在 backend/src/repositories/admin_account_repo.go (Create, GetByUsername, ValidatePassword)
- [X] T021 [P] [US1] 實作 SystemSettings repository 在 backend/src/repositories/settings_repo.go (Get, Set, GetAll)

#### 後端服務層 (US1)

- [X] T022 [US1] 實作認證服務在 backend/src/services/auth_service.go (登入、登出、token 管理)
- [X] T023 [P] [US1] 實作域名服務基礎在 backend/src/services/domain_service.go (ListDomains, GetDomainByID)
- [X] T024 [P] [US1] 實作系統設定服務在 backend/src/services/settings_service.go (GetSettings, UpdateSettings)

#### API 端點 (US1)

- [X] T025 [US1] 實作認證 API 在 backend/src/api/auth.go (POST /api/v1/auth/login, POST /api/v1/auth/logout)
- [X] T026 [P] [US1] 實作域名列表 API 在 backend/src/api/domains.go (GET /api/v1/domains)
- [X] T027 [P] [US1] 實作系統設定 API 在 backend/src/api/settings.go (GET /api/v1/settings, PATCH /api/v1/settings)
- [X] T028 [P] [US1] 實作健康檢查 API 在 backend/src/api/health.go (GET /api/v1/diagnostics/health)

#### 前端頁面 (US1)

- [X] T029 [P] [US1] 實作登入頁面在 frontend/src/pages/login.html (表單、HTMX 整合)
- [X] T030 [P] [US1] 實作歡迎頁面在 frontend/src/pages/dashboard.html (系統概覽、快速入門指南)
- [X] T031 [P] [US1] 實作導航元件在 frontend/src/components/navigation.html (側邊欄、頂部欄)
- [X] T032 [P] [US1] 實作基礎樣式在 frontend/src/styles/main.css (TailwindCSS 客製化)
- [X] T033 [P] [US1] 實作 API 客戶端在 frontend/src/services/api.js (fetch wrapper、錯誤處理)

#### Helm Chart (US1)

- [X] T034 [US1] 建立 Deployment template 在 helm/domain-manager/templates/deployment.yaml (後端容器、環境變數、健康檢查)
- [X] T035 [P] [US1] 建立 Service template 在 helm/domain-manager/templates/service.yaml (ClusterIP service)
- [X] T036 [P] [US1] 建立 Ingress template 在 helm/domain-manager/templates/ingress.yaml (可選 Ingress 配置)
- [X] T037 [P] [US1] 建立 PersistentVolumeClaim template 在 helm/domain-manager/templates/pvc.yaml (SQLite 資料儲存)
- [X] T038 [P] [US1] 建立 ConfigMap template 在 helm/domain-manager/templates/configmap.yaml (應用配置)
- [X] T039 [P] [US1] 建立 Secret template 在 helm/domain-manager/templates/secret.yaml (預設管理員密碼)
- [X] T040 [P] [US1] 撰寫 Helm values 在 helm/domain-manager/values.yaml (預設配置、可客製化參數)
- [X] T041 [P] [US1] 撰寫 Helm README 在 helm/domain-manager/README.md (安裝指引、配置說明)

**檢查點**: US1 完整功能,可獨立測試 - 在全新叢集執行 `helm install` 並登入介面

---

## Phase 4: 使用者故事 2 - 設定外部域名並自動取得 SSL 憑證 🔒

**目標**: 使用者可以新增域名、自動申請 Let's Encrypt 憑證、上傳自訂憑證

**獨立測試**: 新增域名配置、指向測試服務,從外部存取 https://example.com 並驗證憑證

### 實作任務

#### 資料模型與 Repository (US2)

- [X] T042 [P] [US2] 建立 Certificate 模型在 backend/src/models/certificate.go (struct 定義、憑證解析、驗證)
- [X] T043 [US2] 實作 Certificate repository 在 backend/src/repositories/certificate_repo.go (Create, GetByID, GetByDomain, List, Update, Delete, GetExpiring)

#### K8s 操作層 (US2)

- [X] T044 [US2] 實作 Ingress 管理在 backend/src/k8s/ingress.go (CreateIngress, UpdateIngress, DeleteIngress, GetIngress)
- [X] T045 [P] [US2] 實作 Secret 管理在 backend/src/k8s/secret.go (CreateTLSSecret, UpdateSecret, DeleteSecret)
- [X] T046 [P] [US2] 實作 Service 查詢在 backend/src/k8s/service.go (ListServices, GetService, ValidateService)
- [X] T047 [P] [US2] 實作 Service 健康檢查在 backend/src/k8s/health.go (CheckServiceHealth, GetEndpoints)

#### SSL 憑證管理 (US2)

- [ ] T048 [US2] 實作 Let's Encrypt 客戶端在 backend/src/services/letsencrypt/client.go (ACME 客戶端初始化、帳戶管理)
- [ ] T049 [US2] 實作憑證申請在 backend/src/services/letsencrypt/obtain.go (HTTP-01 challenge、憑證取得)
- [ ] T050 [US2] 實作憑證續約在 backend/src/services/letsencrypt/renew.go (自動續約邏輯、錯誤處理)
- [ ] T051 [P] [US2] 實作憑證驗證在 backend/src/services/certificate/validation.go (PEM 格式驗證、私鑰匹配、到期檢查)
- [ ] T052 [P] [US2] 實作憑證加密在 backend/src/services/certificate/encryption.go (AES-256-GCM 加密私鑰)

#### 後端服務層 (US2)

- [X] T053 [US2] 擴展域名服務在 backend/src/services/domain_service.go (CreateDomain, UpdateDomain, DeleteDomain, GetDomainStatus) - 已整合 K8s Ingress 操作
- [X] T054 [US2] 實作憑證服務在 backend/src/services/certificate_service.go (UploadCertificate, GetCertificate, ListCertificates, DeleteCertificate) - 核心功能完成，暫不含 Let's Encrypt
- [X] T055 [US2] 實作診斷服務在 backend/src/services/diagnostic_service.go (CheckDNS, CheckIngress, LogDiagnostic)
- [ ] T056 [P] [US2] 實作後台任務調度在 backend/src/services/scheduler/scheduler.go (憑證續約排程、健康檢查) - 延後實作
- [ ] T057 [P] [US2] 實作後台任務: 憑證監控在 backend/src/services/scheduler/cert_monitor.go (定期檢查到期憑證) - 延後實作

#### API 端點 (US2)

- [X] T058 [US2] 實作域名建立 API 在 backend/src/api/domains.go (POST /api/v1/domains) - Phase 3 已完成
- [X] T059 [P] [US2] 實作域名詳情 API 在 backend/src/api/domains.go (GET /api/v1/domains/{id}) - Phase 3 已完成
- [X] T060 [P] [US2] 實作域名更新 API 在 backend/src/api/domains.go (PUT /api/v1/domains/{id}) - Phase 3 已完成
- [X] T061 [P] [US2] 實作域名刪除 API 在 backend/src/api/domains.go (DELETE /api/v1/domains/{id}) - Phase 3 已完成
- [X] T062 [P] [US2] 實作域名狀態 API 在 backend/src/api/domains.go (GET /api/v1/domains/{id}/status) - Phase 3 已完成
- [ ] T063 [P] [US2] 實作域名診斷 API 在 backend/src/api/domains.go (GET /api/v1/domains/{id}/diagnostics)
- [X] T064 [US2] 實作憑證列表 API 在 backend/src/api/certificates.go (GET /api/v1/certificates)
- [X] T065 [P] [US2] 實作憑證上傳 API 在 backend/src/api/certificates.go (POST /api/v1/certificates)
- [X] T066 [P] [US2] 實作憑證詳情 API 在 backend/src/api/certificates.go (GET /api/v1/certificates/{id})
- [X] T067 [P] [US2] 實作憑證刪除 API 在 backend/src/api/certificates.go (DELETE /api/v1/certificates/{id})
- [ ] T068 [P] [US2] 實作憑證續約 API 在 backend/src/api/certificates.go (POST /api/v1/certificates/{id}/renew) - 延後實作
- [ ] T069 [P] [US2] 實作到期憑證列表 API 在 backend/src/api/certificates.go (GET /api/v1/certificates/expiring)
- [X] T070 [US2] 實作服務發現 API 在 backend/src/api/services.go (GET /api/v1/services, GET /api/v1/services/{namespace}/{name})
- [X] T071 [P] [US2] 實作健康檢查 API 在 backend/src/api/handlers.go (GET /api/v1/diagnostics/health)
- [X] T072 [P] [US2] 實作診斷日誌 API 在 backend/src/api/handlers.go (GET /api/v1/diagnostics/logs)

#### 前端頁面與元件 (US2)

- [X] T073 [P] [US2] 實作域名列表頁面在 frontend/src/pages/domains.html (表格、篩選、分頁)
- [X] T074 [P] [US2] 實作新增域名表單在 frontend/src/components/domain-form.html (HTMX 動態表單、服務選擇器)
- [X] T075 [P] [US2] 實作域名詳情頁面在 frontend/src/pages/domain-detail.html (狀態顯示、診斷資訊)
- [X] T076 [P] [US2] 實作憑證列表頁面在 frontend/src/pages/certificates.html (表格、到期提示)
- [X] T077 [P] [US2] 實作憑證上傳表單在 frontend/src/pages/certificates.html (模態對話框整合、PEM 格式驗證)

**檢查點**: US2 完整功能,可獨立測試 - 新增域名並驗證 SSL 憑證

---

## Phase 4.5: CI/CD 與容器化 🚀

**目的**: 建立完整的 CI/CD 基礎設施，為 K8s 部署做準備

**策略**: 在進入更多功能開發前，先建立自動化部署流程

### 實作任務

#### 容器化 (Phase 4.5)

- [X] T133 [P] [Phase 4.5] 建立根目錄 Dockerfile 在 Dockerfile (多階段構建: Frontend + Backend)
- [X] T134 [P] [Phase 4.5] 優化 Backend Dockerfile 在 backend/Dockerfile (CGO 支援、非 root 使用者、健康檢查)
- [X] T135 [P] [Phase 4.5] 建立 .dockerignore 在 .dockerignore (優化建置速度)
- [X] T136 [P] [Phase 4.5] 建立 Docker Compose 配置在 docker-compose.yml (本地測試環境、Mock 模式)

#### CI/CD Pipeline (Phase 4.5)

- [X] T137 [Phase 4.5] 建立 GitHub Actions CI workflow 在 .github/workflows/ci.yml (測試、Lint、Docker 建置)
- [X] T138 [Phase 4.5] 建立 GitHub Actions CD workflow 在 .github/workflows/cd.yml (自動部署到 Staging/Production)

#### 文件與工具 (Phase 4.5)

- [X] T139 [P] [Phase 4.5] 建立 Docker 清理腳本在 docker-cleanup.sh (一鍵恢復系統)
- [X] T140 [P] [Phase 4.5] 撰寫 Docker 使用指南在 DOCKER.md (完整操作文件、疑難排解)

**檢查點**: CI/CD 完整功能,可獨立測試 - Docker 建置成功、容器運行正常、完全清理成功

**測試結果**:
- ✅ Docker 建置: 成功 (~70秒)
- ✅ 容器啟動: 成功
- ✅ API 測試: 登入成功
- ✅ 完全清理: 無殘留資源

---

## Phase 5: 使用者故事 3 - 管理子域名並分派到不同服務 🌐

**目標**: 使用者可以新增子域名、每個子域名獨立配置憑證

**獨立測試**: 新增多個子域名,每個指向不同服務,分別存取驗證路由

### 實作任務

#### 後端增強 (US3)

- [ ] T078 [P] [US3] 實作子域名驗證在 backend/src/services/domain_service.go (CheckSubdomainConflict, ValidateSubdomain)
- [ ] T079 [P] [US3] 實作萬用字元憑證支援在 backend/src/services/certificate_service.go (WildcardCertificate detection)
- [ ] T080 [P] [US3] 實作域名分組邏輯在 backend/src/services/domain_service.go (GroupByRootDomain, GetSubdomains)
- [ ] T081 [P] [US3] 實作批次域名操作在 backend/src/services/domain_service.go (BulkCreate, BulkDelete, BulkUpdate)

#### API 端點 (US3)

- [ ] T082 [P] [US3] 實作子域名列表 API 在 backend/src/api/domains.go (GET /api/v1/domains?parent={domain})
- [ ] T083 [P] [US3] 實作域名樹狀結構 API 在 backend/src/api/domains.go (GET /api/v1/domains/tree)
- [ ] T084 [P] [US3] 實作批次域名建立 API 在 backend/src/api/domains.go (POST /api/v1/domains/batch)

#### 前端增強 (US3)

- [ ] T085 [P] [US3] 實作子域名樹狀檢視在 frontend/src/components/domain-tree.html (Alpine.js 折疊樹)
- [ ] T086 [P] [US3] 實作快速新增子域名在 frontend/src/components/subdomain-quick-add.html (模態對話框、快速表單)
- [ ] T087 [P] [US3] 實作域名篩選與搜尋在 frontend/src/components/domain-filter.html (即時搜尋、多條件篩選)
- [ ] T088 [P] [US3] 實作域名批次操作在 frontend/src/components/domain-bulk-actions.html (批次刪除、批次啟用/停用)

#### 資料模型增強 (US3)

- [ ] T089 [P] [US3] 建立 DiagnosticLog 模型在 backend/src/models/diagnostic_log.go
- [ ] T090 [US3] 實作 DiagnosticLog repository 在 backend/src/repositories/diagnostic_log_repo.go (Create, List, MarkResolved)

#### 診斷增強 (US3)

- [ ] T091 [P] [US3] 實作子域名健康檢查在 backend/src/services/diagnostic_service.go (CheckSubdomainHealth, CheckAllSubdomains)
- [ ] T092 [P] [US3] 實作診斷日誌前端頁面在 frontend/src/pages/diagnostics.html (日誌列表、篩選、標記已解決)

#### Helm 增強 (US3)

- [ ] T093 [P] [US3] 更新 Helm values 在 helm/domain-manager/values.yaml (子域名預設配置、Ingress class 設定)
- [ ] T094 [P] [US3] 更新 Helm README 在 helm/domain-manager/README.md (子域名管理說明)
- [ ] T095 [P] [US3] 新增 Helm examples 在 helm/domain-manager/examples/ (子域名配置範例)

**檢查點**: US3 完整功能,可獨立測試 - 新增多個子域名並驗證

---

## Phase 6: 使用者故事 4 - 透過 API 或 MCP 程式化操作 🤖

**目標**: 開發者和 AI 助理可以透過 REST API 和 MCP 管理域名

**獨立測試**: 使用 curl 呼叫 API,使用 Claude Desktop 透過 MCP 操作

### 實作任務

#### API 金鑰管理 (US4)

- [ ] T096 [P] [US4] 建立 APIKey 模型在 backend/src/models/api_key.go
- [ ] T097 [US4] 實作 APIKey repository 在 backend/src/repositories/api_key_repo.go (Create, GetByKey, List, Delete, UpdateLastUsed)
- [ ] T098 [US4] 實作 API 金鑰服務在 backend/src/services/api_key_service.go (GenerateKey, ValidateKey, RevokeKey)
- [ ] T099 [US4] 實作 API 金鑰管理 API 在 backend/src/api/api_keys.go (GET /api/v1/api-keys, POST /api/v1/api-keys, DELETE /api/v1/api-keys/{id})
- [ ] T100 [P] [US4] 實作 API 金鑰管理頁面在 frontend/src/pages/api-keys.html (列表、建立、刪除)

#### 備份功能 (US4)

- [ ] T101 [US4] 實作資料庫備份服務在 backend/src/services/backup_service.go (CreateBackup, ListBackups, CleanOldBackups)
- [ ] T102 [US4] 實作備份 API 在 backend/src/api/backup.go (POST /api/v1/backup, GET /api/v1/backup)
- [ ] T103 [P] [US4] 實作備份管理頁面在 frontend/src/pages/backup.html (建立備份、下載備份)

#### MCP 伺服器實作 (US4)

- [ ] T104 [US4] 實作 MCP 基礎結構在 backend/src/mcp/server.go (JSON-RPC 2.0 handler、路由)
- [ ] T105 [P] [US4] 實作 MCP 錯誤處理在 backend/src/mcp/errors.go (RPCError 定義、錯誤碼映射)
- [ ] T106 [US4] 實作 MCP Tools: 域名管理在 backend/src/mcp/tools_domains.go (list_domains, get_domain, create_domain, update_domain, delete_domain)
- [ ] T107 [P] [US4] 實作 MCP Tools: 服務發現在 backend/src/mcp/tools_services.go (list_services)
- [ ] T108 [P] [US4] 實作 MCP Tools: 憑證管理在 backend/src/mcp/tools_certificates.go (get_certificate_status, list_expiring_certificates, renew_certificate)
- [ ] T109 [P] [US4] 實作 MCP Tools: 診斷在 backend/src/mcp/tools_diagnostics.go (check_dns, get_diagnostics, get_system_health)
- [ ] T110 [US4] 實作 MCP Resources: 域名在 backend/src/mcp/resources_domains.go (domain://list, domain://{name})
- [ ] T111 [P] [US4] 實作 MCP Resources: 服務在 backend/src/mcp/resources_services.go (service://list)
- [ ] T112 [P] [US4] 實作 MCP Resources: 憑證在 backend/src/mcp/resources_certificates.go (certificate://list)
- [ ] T113 [P] [US4] 實作 MCP Resources: 診斷在 backend/src/mcp/resources_diagnostics.go (diagnostics://logs)
- [ ] T114 [US4] 實作 MCP 路由在 backend/src/api/mcp.go (POST /mcp)

#### MCP 文件與測試 (US4)

- [ ] T115 [P] [US4] 撰寫 MCP 客戶端範例在 docs/mcp-examples.md (Claude Desktop 設定、使用範例)
- [ ] T116 [P] [US4] 建立 MCP 測試腳本在 backend/tests/mcp/test_mcp_tools.sh (curl 測試所有工具)

#### API 文件與測試 (US4)

- [ ] T117 [P] [US4] 建立 Postman collection 在 docs/postman/domain-manager.json (所有 API 端點範例)
- [ ] T118 [P] [US4] 撰寫 API 使用文件在 docs/api-usage.md (認證、端點說明、範例)

#### Helm 增強 (US4)

- [ ] T119 [P] [US4] 更新 Helm templates 在 helm/domain-manager/templates/ (MCP 端點配置、API 金鑰 Secret)
- [ ] T120 [P] [US4] 更新 Helm values 在 helm/domain-manager/values.yaml (API 和 MCP 配置選項)

#### 前端增強 (US4)

- [ ] T121 [P] [US4] 實作 API 文件頁面在 frontend/src/pages/api-docs.html (嵌入式 Swagger UI 或 OpenAPI 檢視)
- [ ] T122 [P] [US4] 實作系統設定頁面在 frontend/src/pages/settings.html (Let's Encrypt 郵箱、Ingress class、續約設定)

#### 整合增強 (US4)

- [ ] T123 [P] [US4] 實作速率限制在 backend/src/middleware/rate_limit.go (API 和 MCP 速率限制)
- [ ] T124 [P] [US4] 實作 CORS 中介軟體在 backend/src/middleware/cors.go (跨域請求支援)
- [ ] T125 [P] [US4] 實作 API 版本控制在 backend/src/api/versioning.go (v1 路由命名空間)
- [ ] T126 [P] [US4] 實作請求追蹤在 backend/src/middleware/tracing.go (Request ID、分散式追蹤)

**檢查點**: US4 完整功能,可獨立測試 - 透過 API 和 MCP 操作域名

---

## Phase 7: 收尾與跨功能關注點

**目的**: 影響多個使用者故事的改進和文件

- [ ] T127 [P] 撰寫專案 README 在 README.md (專案簡介、功能特色、快速開始、架構說明)
- [ ] T128 [P] 建立快速入門文件在 docs/quickstart.md (從零開始的完整部署與配置指南)
- [ ] T129 [P] 建立故障排除文件在 docs/troubleshooting.md (常見問題、診斷步驟、解決方案)
- [ ] T130 [P] 建立架構文件在 docs/architecture.md (系統架構圖、元件說明、資料流)
- [ ] T131 執行完整端對端驗證在全新 K8s 叢集 (遵循 docs/quickstart.md 步驟驗證)
- [ ] T132 程式碼審查與重構 (檢查所有 TODO、優化效能、統一程式碼風格)

---

## 任務依賴關係

### 阻塞依賴 (必須按順序執行)

1. **Phase 1 → Phase 2**: 專案結構必須先建立
2. **Phase 2 → Phase 3/4/5/6**: 基礎設施必須先完成
3. **T006 → T008**: Schema 必須先定義
4. **T016/T017/T018 → T019/T020/T021**: 模型必須先定義
5. **T019 → T023**: Repository 必須先完成
6. **T022/T023 → T025/T026**: 服務層必須先完成
7. **T034-T041 → T131**: Helm Chart 必須先完成
8. **T096/T097 → T098**: API Key 模型和 Repository 必須先完成
9. **T104 → T106-T114**: MCP 基礎結構必須先完成

### 平行執行 (標記 [P])

所有標記 [P] 的任務可以在其依賴完成後平行執行。例如:
- Phase 2 完成後,US1/US2/US3/US4 的模型定義 (標記 [P]) 可以同時開發
- 前端頁面 (T029-T033) 可以在 API 端點完成後平行開發
- Helm templates (T035-T041) 可以在 Deployment template 完成後平行開發

---

## MVP 範圍建議

**最小可行產品 (MVP)** 應該包含:

### Phase 1-3 (US1) + 部分 Phase 4 (US2 核心)

- ✅ Phase 1: 專案設定
- ✅ Phase 2: 基礎設施
- ✅ Phase 3: US1 完整 (Helm 部署 + Web 介面)
- ✅ Phase 4 部分: T042-T062, T073-T075 (域名管理 + Let's Encrypt 自動憑證)

**MVP 任務數**: 約 75 個任務

**預估開發時間**: 2-3 週 (1 位全職開發者)

---

## 使用者故事任務分布

| 使用者故事 | 任務範圍 | 任務數 | 關鍵功能 |
|---------|---------|-------|---------|
| **US1** (P1) | T016-T041 | 26 | Helm 部署 + Web 介面 + 登入 |
| **US2** (P2) | T042-T077 | 36 | 域名管理 + Let's Encrypt + 自訂憑證 |
| **US3** (P3) | T078-T095 | 18 | 子域名 + 診斷 + 批次操作 |
| **US4** (P4) | T096-T126 | 31 | API 金鑰 + MCP + 備份 |
| **跨功能** | T127-T132 | 6 | 文件 + 驗證 + 重構 |

---

## 開發建議

### 開發流程

1. **Phase 1-2 (基礎)**: 1-2 天 - 建立專案結構和基礎設施
2. **Phase 3 (US1)**: 3-5 天 - 實作 MVP 核心功能
3. **Phase 4 (US2)**: 5-7 天 - SSL 憑證自動化
4. **Phase 5 (US3)**: 2-3 天 - 子域名管理
5. **Phase 6 (US4)**: 4-5 天 - API 和 MCP
6. **Phase 7 (收尾)**: 1-2 天 - 文件和驗證

### 並行開發策略

如果有多位開發者:
- **開發者 A**: Phase 1-2 → Phase 3 後端 → Phase 4 後端
- **開發者 B**: Phase 3 前端 → Phase 4 前端 → Phase 5
- **開發者 C**: Phase 3 Helm → Phase 6 MCP → Phase 7 文件

### 測試策略

每個 Phase 完成後執行檢查點測試:
- **Phase 3 檢查點**: `helm install` 測試
- **Phase 4 檢查點**: 域名建立和 SSL 測試
- **Phase 5 檢查點**: 子域名測試
- **Phase 6 檢查點**: API 和 MCP 測試
- **Phase 7 檢查點**: 完整端對端測試

---

**生成日期**: 2025-11-07
**任務總數**: 132 個
**預估總時間**: 3-4 週 (1 位全職開發者) 或 1.5-2 週 (2-3 位開發者平行開發)
