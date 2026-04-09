# Implementation Plan: Cloudflare DNS + cert-manager DNS-01 整合

**Branch**: `003-cloudflare-dns01` | **Date**: 2026-04-09 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/003-cloudflare-dns01/spec.md`

## Summary

整合 Cloudflare 免費 DNS 和 cert-manager DNS-01 solver，實現全免費的 wildcard 憑證自動申請。使用者在 UI 設定 Cloudflare API Token 後，系統自動建立 K8s Secret 和 ClusterIssuer，cert-manager 透過 Cloudflare DNS API 完成 DNS-01 challenge。

## Technical Context

**Language/Version**: Go 1.22+
**Primary Dependencies**: go-chi/chi v5, client-go, cert-manager (Helm dependency)
**Storage**: SQLite (system_settings) + K8s Secret (token) + K8s ClusterIssuer
**Testing**: go test, curl, mock mode
**Target Platform**: Kubernetes 1.20+ (含 K3s)
**Project Type**: web (backend Go + frontend HTMX)
**Performance Goals**: Token 設定到 ClusterIssuer 就緒在 30 秒內
**Constraints**: 僅支援 Cloudflare 免費 DNS，不支援其他 DNS provider
**Scale/Scope**: 新增 Cloudflare service + 修改 Helm templates + 前端設定 UI

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原則 | 狀態 | 說明 |
|------|------|------|
| 簡單優先 | ✅ 通過 | 利用 cert-manager 原生 DNS-01 solver，不自行實作 ACME |
| 清晰的資料模型 | ✅ 通過 | token 存 settings + Secret，ClusterIssuer 由 Helm/K8s 管理 |
| 適當的測試覆蓋 | ✅ 通過 | mock 模式驗證 Secret/ClusterIssuer 建立，實際 DNS-01 需 E2E |
| 文件化決策 | ✅ 通過 | 決策記錄在 research.md |

## Project Structure

### Documentation (this feature)

```text
specs/003-cloudflare-dns01/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
└── tasks.md
```

### Source Code (repository root)

```text
backend/
├── src/
│   ├── services/
│   │   └── cloudflare_service.go  # 新增：token 驗證、K8s Secret 管理、ClusterIssuer 管理
│   ├── api/
│   │   └── handlers.go            # 修改：新增 Cloudflare 設定 API handlers
│   └── k8s/
│       └── certmanager.go         # 新增：ClusterIssuer CRUD 操作

frontend/
└── src/
    └── pages/
        └── settings.html          # 修改：新增 Cloudflare 設定區塊

helm/
└── domain-manager/
    ├── templates/
    │   ├── clusterissuer.yaml     # 修改：加入 DNS-01 solver
    │   └── secret-cloudflare.yaml # 新增：Cloudflare token Secret
    └── values.yaml                # 修改：新增 cloudflare 區塊
```

**Structure Decision**: Cloudflare 相關邏輯集中在 `cloudflare_service.go`，K8s 操作（Secret、ClusterIssuer）在 `k8s/certmanager.go`。遵循既有分層慣例。
