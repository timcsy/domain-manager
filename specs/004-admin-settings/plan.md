# Implementation Plan: 管理員設定整合

**Branch**: `004-admin-settings` | **Date**: 2026-04-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-admin-settings/spec.md`

## Summary

在系統設定頁面整合管理員帳號管理：密碼修改（驗證舊密碼 + 強制重新登入）、email 修改、Cloudflare Token 直接更新。擴展既有 AdminAccount repository 和前端 settings.html。

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: go-chi/chi v5, bcrypt (golang.org/x/crypto)
**Storage**: SQLite — admin_accounts 表已有 username, password_hash, email 欄位
**Testing**: go test, mock mode
**Target Platform**: Kubernetes 1.20+
**Project Type**: web (backend Go + frontend HTMX)
**Performance Goals**: 密碼修改操作在 2 秒內完成
**Constraints**: 單一管理員帳號，不需多帳號管理
**Scale/Scope**: 擴展 3 個既有檔案 + 修改前端

## Constitution Check

| 原則 | 狀態 | 說明 |
|------|------|------|
| 簡單優先 | ✅ | 擴展既有 repo/service，不建立新模組 |
| 清晰的資料模型 | ✅ | 使用既有 admin_accounts 表，不新增欄位 |
| 適當的測試覆蓋 | ✅ | bcrypt 驗證 + session 清除可驗證 |
| 文件化決策 | ✅ | 記錄在 research.md |

## Project Structure

### Source Code

```text
backend/
├── src/
│   ├── repositories/
│   │   └── admin_account_repo.go  # 修改：新增 UpdatePassword, UpdateEmail
│   ├── services/
│   │   └── auth_service.go        # 修改：新增 ChangePassword, UpdateEmail
│   └── api/
│       └── handlers.go            # 修改：新增 admin 帳號 handlers

frontend/
└── src/
    └── pages/
        └── settings.html          # 修改：新增帳號管理區塊
```

**Structure Decision**: 全部在既有檔案中擴展，遵循「沿用既有檔案」經驗教訓。
