# Implementation Plan: 多 Ingress Controller 支援

**Branch**: `002-multi-ingress` | **Date**: 2026-04-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-multi-ingress/spec.md`

## Summary

移除程式碼中硬編碼的 `"nginx"` ingress class，改為從系統設定動態讀取。建立 Ingress Controller profile 機制，根據 controller 類型（nginx/traefik）自動套用對應的 annotation。

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: go-chi/chi v5, client-go, modernc.org/sqlite
**Storage**: SQLite (WAL mode) — `system_settings` 表已有 `default_ingress_class` 和 `ingress_annotations` 欄位
**Testing**: go test, curl 手動測試
**Target Platform**: Kubernetes 1.20+ (含 K3s)
**Project Type**: web (backend Go + frontend HTMX)
**Performance Goals**: Ingress 建立/更新操作在 5 秒內完成
**Constraints**: 單一實例部署，僅使用標準 K8s Ingress 資源（不使用 CRD）
**Scale/Scope**: 影響 3 個 Go 檔案的硬編碼替換 + 1 個新的 ingress profile 機制

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原則 | 狀態 | 說明 |
|------|------|------|
| 簡單優先 | ✅ 通過 | 使用 map 儲存 controller profile，不過度抽象 |
| 清晰的資料模型 | ✅ 通過 | 重用既有 `system_settings` 表，不新增 DB 表 |
| 適當的測試覆蓋 | ✅ 通過 | 透過 go build/vet + mock 模式手動測試驗證 |
| 文件化決策 | ✅ 通過 | 決策記錄在 research.md |

## Project Structure

### Documentation (this feature)

```text
specs/002-multi-ingress/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
backend/
├── src/
│   ├── k8s/
│   │   ├── ingress.go          # 修改：讀取動態 ingress class + annotations
│   │   └── ingress_profiles.go # 新增：controller profile 定義
│   ├── services/
│   │   ├── domain_service.go       # 修改：移除硬編碼 nginx
│   │   └── certificate_service.go  # 修改：移除硬編碼 nginx
│   └── api/
│       └── handlers.go         # 微調：確保 settings 正確傳遞
├── database/
│   └── migrations/             # 無變更（schema 已有 default_ingress_class）

frontend/
└── src/
    └── pages/
        └── settings.html       # 微調：ingress class 改為下拉選單

helm/
└── domain-manager/
    └── values.yaml             # 確認 defaultIngressClass 正確傳入
```

**Structure Decision**: 在既有 `k8s/` 目錄新增 `ingress_profiles.go` 存放 controller profile 定義，遵循經驗教訓「遵循既有程式碼的組織慣例」。
