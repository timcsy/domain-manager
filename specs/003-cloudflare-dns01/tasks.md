# Tasks: Cloudflare DNS + cert-manager DNS-01 整合

**Input**: Design documents from `/specs/003-cloudflare-dns01/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: 無明確要求。透過 mock 模式 + 實際叢集 E2E 驗證。

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, US3, US4)

---

## Phase 1: Setup

**Purpose**: 建立 Cloudflare 和 cert-manager 相關的基礎模組

- [x] T001 新增 cloudflare_enabled 和 cloudflare_api_token 設定到 backend/database/migrations/001_init.up.sql 的 system_settings 預設值
- [x] T002 [P] 建立 backend/src/k8s/certmanager.go（ClusterIssuer 和 Secret 的 CRUD 操作，含 mock 模式支援）
- [x] T003 [P] 建立 backend/src/services/cloudflare_service.go 基礎結構（struct、constructor、依賴注入）

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 路由註冊和服務初始化

**⚠️ CRITICAL**: US1-US4 都依賴此階段完成

- [x] T004 在 backend/src/api/handlers.go 新增 cloudflareService 全域變數並在 InitializeServices 中初始化
- [x] T005 在 backend/src/api/routes.go 新增 /api/v1/cloudflare 路由組（token POST/DELETE、status GET）

**Checkpoint**: 路由和服務就緒，user story 實作可以開始

---

## Phase 3: User Story 1 - 設定 Cloudflare API Token (Priority: P1) 🎯 MVP

**Goal**: 使用者能在 UI 設定 Cloudflare token，系統驗證後儲存並建立 K8s Secret

**Independent Test**: 在 UI 輸入 token，確認驗證成功、settings 儲存、K8s Secret 建立

### Implementation for User Story 1

- [x] T006 [US1] 實作 backend/src/services/cloudflare_service.go 的 ValidateToken 方法（呼叫 Cloudflare API v4 /user/tokens/verify）
- [x] T007 [US1] 實作 backend/src/services/cloudflare_service.go 的 SaveToken 方法（儲存到 system_settings + 呼叫 k8s 建立 Secret）
- [x] T008 [US1] 實作 backend/src/services/cloudflare_service.go 的 GetStatus 方法（回傳 enabled、token_set、token_valid 狀態）
- [x] T009 [US1] 實作 backend/src/services/cloudflare_service.go 的 RemoveToken 方法（清除 settings + 刪除 K8s Secret）
- [x] T010 [US1] 實作 backend/src/k8s/certmanager.go 的 CreateOrUpdateCloudflareSecret 方法（在 cert-manager namespace 建立 Opaque Secret）
- [x] T011 [US1] 實作 backend/src/k8s/certmanager.go 的 DeleteCloudflareSecret 方法
- [x] T012 [US1] 實作 backend/src/api/handlers.go 的 HandleSetCloudflareToken（POST /api/v1/cloudflare/token）
- [x] T013 [US1] 實作 backend/src/api/handlers.go 的 HandleGetCloudflareStatus（GET /api/v1/cloudflare/status）
- [x] T014 [US1] 實作 backend/src/api/handlers.go 的 HandleDeleteCloudflareToken（DELETE /api/v1/cloudflare/token）
- [x] T015 [US1] 更新 frontend/src/pages/settings.html，新增 Cloudflare 設定區塊（token 輸入、驗證按鈕、狀態顯示）
- [x] T016 [US1] 驗證 go build + go vet 通過

**Checkpoint**: US1 完成 — token 設定和驗證完整可用

---

## Phase 4: User Story 2 - 自動建立 DNS-01 ClusterIssuer (Priority: P2)

**Goal**: token 設定後自動建立含 DNS-01 solver 的 ClusterIssuer

**Independent Test**: 設定 token 後檢查 ClusterIssuer 包含 DNS-01 cloudflare solver

### Implementation for User Story 2

- [x] T017 [US2] 實作 backend/src/k8s/certmanager.go 的 CreateOrUpdateClusterIssuer 方法（建立含 DNS-01 + HTTP-01 雙 solver 的 ClusterIssuer）
- [x] T018 [US2] 實作 backend/src/k8s/certmanager.go 的 GetClusterIssuerStatus 方法（檢查 ClusterIssuer 是否 ready）
- [x] T019 [US2] 修改 backend/src/services/cloudflare_service.go 的 SaveToken，儲存 token 後自動呼叫 CreateOrUpdateClusterIssuer
- [x] T020 [US2] 修改 backend/src/services/cloudflare_service.go 的 RemoveToken，移除 token 後將 ClusterIssuer 降級為僅 HTTP-01
- [x] T021 [US2] 更新 backend/src/services/cloudflare_service.go 的 GetStatus，加入 cluster_issuer_ready 狀態
- [x] T022 [US2] 驗證 go build + go vet 通過

**Checkpoint**: US2 完成 — ClusterIssuer 自動管理

---

## Phase 5: User Story 3 - 申請 Wildcard 憑證 (Priority: P3)

**Goal**: 建立 wildcard 域名時透過 DNS-01 自動申請 wildcard 憑證

**Independent Test**: 建立 `*.example.com` 域名，cert-manager 透過 DNS-01 申請憑證成功

### Implementation for User Story 3

- [x] T023 [US3] 修改 backend/src/services/domain_service.go，建立域名時判斷 wildcard 模式，設定 cert-manager annotation 使用 DNS-01 ClusterIssuer
- [x] T024 [US3] 修改 backend/src/services/domain_service.go，wildcard 域名的 Ingress TLS 使用共用的 wildcard Secret 名稱
- [x] T025 [US3] 修改 frontend/src/pages/domains.html 或域名建立表單，新增 wildcard SSL 選項（僅在 Cloudflare 啟用時顯示）
- [x] T026 [US3] 驗證 go build + go vet 通過

**Checkpoint**: US3 完成 — wildcard 憑證可申請，子網域共用

---

## Phase 6: User Story 4 - Helm 預設 Cloudflare 配置 (Priority: P4)

**Goal**: Helm 部署時可預設 Cloudflare token，自動完成所有配置

**Independent Test**: 使用 --set cloudflare.apiToken=xxx 部署，Secret 和 ClusterIssuer 自動建立

### Implementation for User Story 4

- [x] T027 [US4] 更新 helm/domain-manager/values.yaml，新增 cloudflare 區塊（enabled、apiToken）
- [x] T028 [US4] 建立 helm/domain-manager/templates/secret-cloudflare.yaml（條件式建立 Cloudflare token Secret）
- [x] T029 [US4] 修改 helm/domain-manager/templates/clusterissuer.yaml，支援 DNS-01 solver（條件式，當 cloudflare.enabled 時加入）
- [x] T030 [US4] 更新 helm/domain-manager/templates/deployment.yaml，傳入 CLOUDFLARE_ENABLED 和 CLOUDFLARE_API_TOKEN 環境變數
- [x] T031 [US4] 修改 backend/src/api/handlers.go InitializeServices，啟動時讀取 CLOUDFLARE 環境變數並初始化

**Checkpoint**: US4 完成 — Helm 一鍵部署 Cloudflare 整合

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: 收尾、文件、驗證

- [x] T032 [P] 更新 frontend/src/components/sidebar.html，在系統設定連結旁顯示 Cloudflare 狀態指示
- [x] T033 [P] 更新 docs/api-usage.md，加入 Cloudflare API 端點說明
- [x] T034 [P] 更新 README.md，加入 Cloudflare DNS-01 功能說明
- [x] T035 更新 knowledge/vision.md，標記 Cloudflare 整合里程碑完成
- [x] T036 更新 specs/003-cloudflare-dns01/tasks.md 任務狀態

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: 無依賴
- **Phase 2 (Foundational)**: 依賴 Phase 1
- **Phase 3 (US1)**: 依賴 Phase 2 — token 管理是所有功能的前提
- **Phase 4 (US2)**: 依賴 US1 — 需要 token 已儲存才能建立 ClusterIssuer
- **Phase 5 (US3)**: 依賴 US2 — 需要 ClusterIssuer 就緒才能申請 wildcard 憑證
- **Phase 6 (US4)**: 依賴 US1 — Helm 配置需要 token 管理邏輯
- **Phase 7 (Polish)**: 依賴所有 US 完成

### User Story Dependencies

- **US1 (P1)**: Phase 2 完成後即可開始 — 核心 token 管理
- **US2 (P2)**: US1 完成後開始 — 需要 token 才能建立 ClusterIssuer
- **US3 (P3)**: US2 完成後開始 — 需要 DNS-01 ClusterIssuer
- **US4 (P4)**: US1 完成後即可開始 — 可與 US2/US3 平行

### Parallel Opportunities

- T002 和 T003 可平行執行（不同檔案）
- T006-T009 和 T010-T011 可平行執行（service vs k8s 層）
- US4 可與 US2/US3 平行執行
- T032-T034 可全部平行執行

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1: 基礎模組建立
2. Phase 2: 路由和服務初始化
3. Phase 3: Token 設定和驗證
4. **STOP and VALIDATE**: UI 設定 token，確認儲存和 Secret 建立

### Incremental Delivery

1. US1 → Token 管理（MVP）
2. US2 → ClusterIssuer 自動管理
3. US3 → Wildcard 憑證申請
4. US4 → Helm 預設配置

---

## Notes

- 總任務數: 36
- US1: 11 個任務（token 管理 + UI）
- US2: 6 個任務（ClusterIssuer 管理）
- US3: 4 個任務（wildcard 憑證）
- US4: 5 個任務（Helm 整合）
- Setup/Foundational: 5 個任務
- Polish: 5 個任務
- 需要外部 HTTP 呼叫：Cloudflare API token 驗證（需處理 mock 模式）
