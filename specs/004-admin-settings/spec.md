# Feature Specification: 管理員設定整合

**Feature Branch**: `004-admin-settings`  
**Created**: 2026-04-09  
**Status**: Draft  
**Input**: User description: "Admin account management in settings page"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 修改管理員密碼 (Priority: P1)

管理員進入系統設定頁面，在帳號管理區塊輸入舊密碼和新密碼。系統驗證舊密碼正確後更新密碼，並強制重新登入。

**Why this priority**: 安全性最高優先。預設密碼 `admin` 必須能被修改，否則系統不安全。

**Independent Test**: 使用預設密碼登入，在設定頁面修改密碼，確認舊密碼無法登入、新密碼可以登入。

**Acceptance Scenarios**:

1. **Given** 管理員已登入，**When** 輸入正確的舊密碼和新密碼後提交，**Then** 密碼更新成功，系統登出並重導向登入頁面
2. **Given** 管理員已登入，**When** 輸入錯誤的舊密碼，**Then** 顯示「舊密碼不正確」錯誤，密碼不變
3. **Given** 管理員已登入，**When** 新密碼少於 6 個字元，**Then** 顯示「密碼至少 6 個字元」錯誤
4. **Given** 密碼已修改，**When** 使用舊密碼登入，**Then** 登入失敗
5. **Given** 密碼已修改，**When** 使用新密碼登入，**Then** 登入成功

---

### User Story 2 - 修改管理員 Email (Priority: P2)

管理員在系統設定頁面修改自己的 email 地址。此 email 用於系統通知和 Let's Encrypt 註冊。

**Why this priority**: 實用功能，但不影響系統安全性。

**Independent Test**: 在設定頁面修改 email，確認儲存成功且重新載入後顯示新 email。

**Acceptance Scenarios**:

1. **Given** 管理員已登入，**When** 輸入有效的 email 並儲存，**Then** email 更新成功，頁面顯示新 email
2. **Given** 管理員已登入，**When** 輸入無效的 email 格式，**Then** 顯示格式錯誤提示

---

### User Story 3 - 直接更新 Cloudflare Token (Priority: P3)

管理員在 Cloudflare 設定區塊可以直接輸入新的 token 覆蓋舊的，不需要先移除再重新設定。

**Why this priority**: 改善現有 UX，但現有的移除+重設流程仍然可用。

**Independent Test**: 在已有 token 的狀態下，輸入新 token，確認驗證通過後直接覆蓋。

**Acceptance Scenarios**:

1. **Given** Cloudflare token 已設定，**When** 管理員輸入新 token 並點擊「更新」，**Then** 新 token 驗證通過後覆蓋舊的，狀態保持「已啟用」
2. **Given** Cloudflare token 已設定，**When** 管理員輸入無效的新 token，**Then** 顯示驗證失敗，舊 token 不受影響

---

### Edge Cases

- 管理員修改密碼時瀏覽器 session 過期 → 驗證舊密碼前先檢查 session，過期則重導向登入頁
- 密碼修改和其他設定同時提交 → 密碼修改使用獨立的 API 端點，不影響其他設定儲存
- 新密碼和舊密碼相同 → 允許但不報錯（實務上無害）

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系統 MUST 提供修改管理員密碼的功能，需驗證舊密碼
- **FR-002**: 新密碼 MUST 至少 6 個字元
- **FR-003**: 密碼修改成功後 MUST 清除所有 session 並強制重新登入
- **FR-004**: 系統 MUST 提供修改管理員 email 的功能
- **FR-005**: Email MUST 驗證格式有效性
- **FR-006**: Cloudflare token 更新 MUST 在驗證新 token 通過後才覆蓋舊 token
- **FR-007**: 所有帳號管理操作 MUST 整合在系統設定頁面中

### Key Entities

- **AdminAccount**: 管理員帳號，包含 username、password_hash、email、last_login_at
- **Session**: 使用者 session，密碼修改後需全部清除

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 管理員可在 30 秒內完成密碼修改（含輸入和驗證）
- **SC-002**: 密碼修改後 100% 的舊 session 被清除
- **SC-003**: Cloudflare token 更新操作從 2 步（移除+重設）減少為 1 步
- **SC-004**: 所有帳號管理功能集中在同一個頁面，無需跳轉

## Assumptions

- 系統為單一管理員模式，不需要多管理員帳號管理
- 密碼使用 bcrypt hash 儲存（與現有一致）
- 不需要密碼強度指示器（最低 6 字元即可）
- 不需要 email 驗證流程（直接儲存）
