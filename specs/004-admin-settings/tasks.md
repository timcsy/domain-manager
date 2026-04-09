# Tasks: 管理員設定整合

**Input**: Design documents from `/specs/004-admin-settings/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: 無明確要求。透過 mock 模式手動驗證。

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, US3)

---

## Phase 1: Setup

**Purpose**: 擴展既有 repository 和 service 基礎方法

- [x] T001 [P] 新增 UpdatePassword 方法到 backend/src/repositories/admin_account_repo.go（接受 id 和 newPasswordHash，更新 password_hash 和 updated_at）
- [x] T002 [P] 新增 UpdateEmail 方法到 backend/src/repositories/admin_account_repo.go（接受 id 和 email，更新 email 和 updated_at）
- [x] T003 [P] 新增 GetByID 方法到 backend/src/repositories/admin_account_repo.go（依 ID 取得帳號）

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: 路由註冊

- [x] T004 在 backend/src/api/routes.go 新增 /api/v1/admin 路由組（GET /profile, PATCH /password, PATCH /email）

**Checkpoint**: 路由就緒

---

## Phase 3: User Story 1 - 修改管理員密碼 (Priority: P1) 🎯 MVP

**Goal**: 管理員可修改密碼，修改後強制重新登入

**Independent Test**: 用預設密碼登入，修改密碼，確認舊密碼無法登入、新密碼可以

### Implementation for User Story 1

- [x] T005 [US1] 新增 ChangePassword 方法到 backend/src/services/auth_service.go（驗證舊密碼、檢查新密碼長度 >= 6、bcrypt hash 新密碼、呼叫 repo.UpdatePassword）
- [x] T006 [US1] 實作 HandleChangePassword handler 在 backend/src/api/handlers.go（PATCH /api/v1/admin/password，解析 old_password + new_password，呼叫 ChangePassword，成功回傳提示重新登入）
- [x] T007 [US1] 更新 frontend/src/pages/settings.html，新增「帳號管理」區塊：舊密碼、新密碼、確認新密碼輸入框，修改密碼按鈕，成功後清除 localStorage token 並重導向 /login
- [x] T008 [US1] 驗證 go build + go vet 通過

**Checkpoint**: US1 完成 — 密碼修改可用

---

## Phase 4: User Story 2 - 修改管理員 Email (Priority: P2)

**Goal**: 管理員可修改 email

**Independent Test**: 修改 email，重新載入設定頁面確認顯示新 email

### Implementation for User Story 2

- [x] T009 [US2] 新增 UpdateEmail 方法到 backend/src/services/auth_service.go（驗證 email 格式、呼叫 repo.UpdateEmail）
- [x] T010 [US2] 新增 GetProfile 方法到 backend/src/services/auth_service.go（呼叫 repo.GetByUsername 回傳 username + email）
- [x] T011 [US2] 實作 HandleUpdateEmail handler 在 backend/src/api/handlers.go（PATCH /api/v1/admin/email）
- [x] T012 [US2] 實作 HandleGetProfile handler 在 backend/src/api/handlers.go（GET /api/v1/admin/profile）
- [x] T013 [US2] 更新 frontend/src/pages/settings.html 帳號管理區塊：顯示目前 username（唯讀）和 email（可編輯），載入時呼叫 GET /api/v1/admin/profile
- [x] T014 [US2] 驗證 go build + go vet 通過

**Checkpoint**: US2 完成 — email 修改可用

---

## Phase 5: User Story 3 - 直接更新 Cloudflare Token (Priority: P3)

**Goal**: 已啟用狀態下可直接輸入新 token 覆蓋

**Independent Test**: 在已啟用 Cloudflare 狀態下輸入新 token，確認更新成功

### Implementation for User Story 3

- [x] T015 [US3] 修改 frontend/src/pages/settings.html Cloudflare 區塊：已啟用狀態下也顯示 token 輸入框和「更新 Token」按鈕（呼叫既有 POST /api/v1/cloudflare/token）
- [x] T016 [US3] 驗證 go build + go vet 通過

**Checkpoint**: US3 完成 — Cloudflare token 可直接更新

---

## Phase 6: Polish & Cross-Cutting Concerns

- [x] T017 [P] 更新 docs/api-usage.md，加入 /api/v1/admin 端點說明
- [x] T018 [P] 更新 knowledge/vision.md，標記「管理員設定整合」里程碑為完成
- [x] T019 更新 specs/004-admin-settings/tasks.md 任務狀態

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: 無依賴
- **Phase 2 (Foundational)**: 依賴 Phase 1
- **Phase 3 (US1)**: 依賴 Phase 2
- **Phase 4 (US2)**: 依賴 Phase 2（可與 US1 平行）
- **Phase 5 (US3)**: 無後端依賴（僅前端修改）
- **Phase 6 (Polish)**: 依賴所有 US 完成

### Parallel Opportunities

- T001/T002/T003 可全部平行（同一檔案但不同方法）
- US1 和 US2 的後端部分可平行（不同方法）
- US3 僅前端，可與 US1/US2 平行
- T017/T018 可平行

---

## Implementation Strategy

### MVP First (User Story 1)

1. Phase 1: 擴展 repo 方法
2. Phase 2: 路由註冊
3. Phase 3: 密碼修改
4. **STOP and VALIDATE**: 修改密碼，確認強制重新登入

### Incremental Delivery

1. US1 → 密碼修改（MVP）
2. US2 → Email 修改 + Profile 顯示
3. US3 → Cloudflare Token 直接更新

---

## Notes

- 總任務數: 19
- US1: 4 個任務（密碼修改）
- US2: 6 個任務（Email + Profile）
- US3: 2 個任務（前端修改）
- Setup/Foundational: 4 個任務
- Polish: 3 個任務
- 預估影響檔案: 4 個（admin_account_repo.go, auth_service.go, handlers.go, settings.html）+ docs
