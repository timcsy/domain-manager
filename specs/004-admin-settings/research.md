# Research: 管理員設定整合

## R1: 密碼修改後的 Session 處理

**Decision**: 密碼修改成功後，清除伺服器端所有 session，回傳指示前端清除 localStorage token 並重導向登入頁。

**Rationale**: 目前 session 存在記憶體中（handlers.go 的 Auth middleware 簡化實作），沒有集中式 session store。最安全的做法是回傳 success + 前端負責清除 token 和重導向。

**Alternatives considered**:
- 伺服器端 invalidate 所有 session：目前沒有 session store，無法做到 → 前端處理
- 只 invalidate 當前 session：其他 session 仍能用舊密碼的認證 → 不夠安全，但目前單一管理員且沒有多 session tracking，實際上前端清除就夠用

## R2: 密碼 Hash 方式

**Decision**: 沿用既有的 bcrypt cost 10（與 001_init.up.sql 預設 admin 密碼一致）。

**Rationale**: AdminAccountRepository 已有 `ValidatePassword` 使用 bcrypt.CompareHashAndPassword。新密碼用 bcrypt.GenerateFromPassword 產生即可。

## R3: Cloudflare Token 更新流程

**Decision**: 直接呼叫既有的 `CloudflareService.SaveToken(newToken)`，它已處理驗證 + 儲存 + K8s Secret 更新 + ClusterIssuer 同步。

**Rationale**: SaveToken 已經是冪等操作（create or update），不需要先 remove 再 save。前端 UI 改為在「已啟用」狀態也顯示 token 輸入框。
