# Data Model: 管理員設定整合

## 既有實體（擴展方法，不改 schema）

### admin_accounts

| 欄位 | 類型 | 說明 |
|------|------|------|
| id | INTEGER | PK |
| username | VARCHAR(100) | 使用者名稱 |
| password_hash | VARCHAR(255) | bcrypt hash |
| email | VARCHAR(255) | 聯絡 email |
| last_login_at | TIMESTAMP | 最後登入時間 |
| created_at | TIMESTAMP | 建立時間 |
| updated_at | TIMESTAMP | 更新時間 |

### 需新增的 Repository 方法

| 方法 | 說明 |
|------|------|
| UpdatePassword(id, newPasswordHash) | 更新密碼 hash |
| UpdateEmail(id, email) | 更新 email |
| GetByID(id) | 依 ID 取得帳號（profile 用） |

### 需新增的 Service 方法

| 方法 | 說明 |
|------|------|
| ChangePassword(username, oldPassword, newPassword) | 驗證舊密碼後更新 |
| UpdateEmail(username, email) | 更新 email |
| GetProfile(username) | 取得帳號基本資訊 |
