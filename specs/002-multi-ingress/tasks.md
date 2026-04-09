# Tasks: 多 Ingress Controller 支援

**Input**: Design documents from `/specs/002-multi-ingress/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: 無明確要求，不包含測試任務。透過 go build/vet + mock 模式手動驗證。

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, US3)

---

## Phase 1: Setup

**Purpose**: 建立 Ingress Controller profile 基礎機制

- [x] T001 建立 Ingress Controller profile 定義在 backend/src/k8s/ingress_profiles.go (nginx 和 traefik 的預設 annotation map)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 讓 service 層能存取系統設定中的 ingress class

**⚠️ CRITICAL**: US1-US3 都依賴此階段完成

- [x] T002 在 backend/src/services/domain_service.go 中注入 settingsService 依賴 (新增 SetSettingsService 方法或建構子參數)
- [x] T003 [P] 在 backend/src/services/certificate_service.go 中注入 settingsService 依賴

**Checkpoint**: Service 層可以讀取系統設定，user story 實作可以開始

---

## Phase 3: User Story 1 - 從系統設定切換 Ingress Controller (Priority: P1) 🎯 MVP

**Goal**: 移除硬編碼的 `"nginx"`，改為從系統設定動態讀取 ingress class

**Independent Test**: 修改 `default_ingress_class` 設定為 `traefik`，建立域名，確認 Ingress 使用 `traefik` class

### Implementation for User Story 1

- [x] T004 [US1] 修改 backend/src/services/domain_service.go 的 createIngressForDomain 方法，從 settingsService 讀取 default_ingress_class 替代硬編碼 "nginx"
- [x] T005 [US1] 修改 backend/src/services/domain_service.go 的 updateIngressForDomain 方法，同樣從 settingsService 讀取
- [x] T006 [US1] 修改 backend/src/services/certificate_service.go 的 updateDomainIngressWithCertificate 方法，從 settingsService 讀取
- [x] T007 [US1] 在 backend/src/api/handlers.go 的 InitializeServices 中將 settingsService 注入 domainService 和 certificateService
- [x] T008 [US1] 驗證 go build + go vet 通過，無硬編碼的 "nginx" ingress class 殘留

**Checkpoint**: US1 完成 — 系統使用動態 ingress class，可獨立測試

---

## Phase 4: User Story 2 - 不同 Controller 的 Annotation 自動適配 (Priority: P2)

**Goal**: 根據 ingress class 類型自動套用對應的 annotation

**Independent Test**: 分別設定 nginx 和 traefik，建立帶 SSL 的域名，檢查 annotation 差異

### Implementation for User Story 2

- [x] T009 [US2] 在 backend/src/k8s/ingress_profiles.go 新增 GetAnnotationsForController 函數，根據 controller 名稱回傳對應的 annotation map
- [x] T010 [US2] IngressConfig struct 已有 Annotations map[string]string 欄位（既有）
- [x] T011 [US2] CreateIngress 和 UpdateIngress 已套用 cfg.Annotations 到 Ingress metadata.annotations（既有）
- [x] T012 [US2] 修改 backend/src/services/domain_service.go，建立 Ingress 時讀取系統設定的 ingress_annotations 並與 profile annotation 合併
- [x] T013 [US2] 修改 backend/src/services/certificate_service.go，同樣合併 annotation
- [x] T014 [US2] 驗證 go build + go vet 通過

**Checkpoint**: US2 完成 — Traefik 和 Nginx 各自產生正確的 annotation

---

## Phase 5: User Story 3 - Helm 部署時指定 Ingress Controller (Priority: P3)

**Goal**: Helm values 中的 ingress class 設定正確傳入應用程式

**Independent Test**: 使用 `--set subdomains.defaultIngressClass=traefik` 部署，確認系統設定為 traefik

### Implementation for User Story 3

- [x] T015 [US3] 新增 DEFAULT_INGRESS_CLASS 環境變數到 helm/domain-manager/templates/deployment.yaml
- [x] T016 [US3] 修改 backend/src/api/handlers.go InitializeServices，啟動時讀取 DEFAULT_INGRESS_CLASS 環境變數並寫入系統設定
- [x] T017 [US3] 更新 frontend/src/pages/settings.html，將 Ingress class 輸入欄位改為下拉選單（含 nginx、traefik 選項及自訂輸入）

**Checkpoint**: US3 完成 — Helm 部署自動帶入 ingress class 設定

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 收尾和文件更新

- [x] T018 [P] 使用 grep 確認程式碼中無殘留的硬編碼 "nginx" ingress class（排除 ingress_profiles.go 和 fallback default）
- [x] T019 [P] 更新 knowledge/vision.md，標記「多 Ingress Controller 支援」里程碑為完成
- [x] T020 更新 specs/002-multi-ingress/tasks.md 中的任務狀態

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: 無依賴
- **Phase 2 (Foundational)**: 依賴 Phase 1
- **Phase 3 (US1)**: 依賴 Phase 2 — T002/T003 必須完成
- **Phase 4 (US2)**: 依賴 Phase 3 — 需要動態 ingress class 已實作
- **Phase 5 (US3)**: 依賴 Phase 3 — 需要動態讀取已實作
- **Phase 6 (Polish)**: 依賴所有 US 完成

### User Story Dependencies

- **US1 (P1)**: Phase 2 完成後即可開始 — 核心變更
- **US2 (P2)**: US1 完成後開始 — 擴展 annotation 機制
- **US3 (P3)**: US1 完成後即可開始 — 可與 US2 平行

### Parallel Opportunities

- T002 和 T003 可平行執行（不同檔案）
- T004 和 T006 可平行執行（不同檔案）
- US2 和 US3 可平行執行（US1 完成後）
- T018 和 T019 可平行執行

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1: 建立 profile 定義
2. Phase 2: 注入 settingsService
3. Phase 3: 移除硬編碼，動態讀取
4. **STOP and VALIDATE**: 修改設定為 traefik，建立域名驗證

### Incremental Delivery

1. US1 → 動態 ingress class（MVP）
2. US2 → 正確的 annotation
3. US3 → Helm 整合 + UI 改善

---

## Notes

- 總任務數: 20
- US1: 5 個任務（核心變更）
- US2: 6 個任務（annotation 機制）
- US3: 3 個任務（Helm + UI）
- Setup/Foundational: 3 個任務
- Polish: 3 個任務
- 預估影響檔案: 7 個（3 個 Go service + 1 個 k8s + 1 個 handlers + 1 個前端 + 1 個新檔案）
