# 實作計畫: Kubernetes 域名管理器

**分支**: `001-k8s-domain-manager` | **日期**: 2025-11-07 | **規格**: [spec.md](./spec.md)
**輸入**: 功能規格來自 `/specs/001-k8s-domain-manager/spec.md`

## 摘要

建立一個在 Kubernetes 上運行的域名管理器,提供簡單的 Web 介面讓非 K8s 專家也能輕鬆管理域名到服務的路由和 SSL 憑證。系統使用 SQLite 儲存配置,支援 Let's Encrypt 自動憑證管理,並提供 RESTful API 和 MCP 伺服器供外部工具和 AI 助理操作。

## 技術環境

**語言/版本**: Go 1.21+
**主要依賴**:
- K8s 客戶端: client-go v0.29.0+
- Let's Encrypt: lego v4.15.0+
- Web 框架: Chi v5.0.0+
- SQLite: modernc.org/sqlite v1.28.0+ (CGo-free)
- 前端: HTMX + TailwindCSS + Alpine.js

**儲存**: SQLite 資料庫配合 PersistentVolume
**測試**:
- 單元測試: Go testing + testify
- 整合測試: testcontainers-go + kind
- E2E 測試: Playwright (TypeScript)

**目標平台**: Kubernetes 1.20+ (Linux containers)
**專案類型**: Web 應用 (後端 API + 前端介面)
**效能目標**: 介面回應 <2 秒,支援 100+ 域名,憑證申請 <10 分鐘
**約束**: 資源使用 ≤1 CPU 核心 + 512MB 記憶體,開箱即用(最小化配置)
**規模/範圍**: 單一叢集,100 個域名,單一管理員
**預估資源使用**: 記憶體 30-80MB,CPU 0.1-0.5 核心,容器映像 10-15MB

## 憲法檢查

*關卡: 必須在 Phase 0 研究前通過。Phase 1 設計後重新檢查。*

### 一、簡單優先 ✅
- 使用 SQLite 而非複雜的資料庫系統
- 移除 DNS 伺服器功能,專注於路由和 SSL 管理
- 單一管理員模型,避免複雜的權限系統
- **合規**: 專案遵循 YAGNI 原則,避免過度設計

### 二、清晰的資料模型 ✅
- 定義 6 個核心實體:域名配置、SSL 憑證、服務映射、診斷記錄、管理員帳戶、API 金鑰
- 每個實體職責明確
- **合規**: 資料模型清晰且單一職責

### 三、適當的測試覆蓋 ⚠️
- 需在 Phase 0 研究中確定測試策略
- 專注於核心功能:域名路由、SSL 自動化
- **待驗證**: Phase 1 後確認測試涵蓋範圍

### 四、文件化決策 ✅
- 本計畫文件記錄技術決策
- research.md 將記錄技術選擇理由
- **合規**: 決策過程有文件記錄

**總體評估**: ✅ 通過初步檢查,需在 Phase 0 完成技術選擇後重新評估

## 專案結構

### 文件 (本功能)

```text
specs/001-k8s-domain-manager/
├── plan.md              # 本文件 (/speckit.plan 命令輸出)
├── research.md          # Phase 0 輸出 (/speckit.plan 命令)
├── data-model.md        # Phase 1 輸出 (/speckit.plan 命令)
├── quickstart.md        # Phase 1 輸出 (/speckit.plan 命令)
├── contracts/           # Phase 1 輸出 (/speckit.plan 命令)
└── tasks.md             # Phase 2 輸出 (/speckit.tasks 命令 - 不由 /speckit.plan 建立)
```

### 原始碼 (repository root)

```text
backend/
├── src/
│   ├── models/          # 資料模型 (SQLite ORM)
│   ├── services/        # 業務邏輯 (K8s 操作、SSL 管理)
│   ├── api/             # REST API 端點
│   ├── mcp/             # MCP 伺服器實作
│   └── main.go/py       # 應用程式入口
├── database/
│   └── migrations/      # 資料庫遷移腳本
└── tests/
    ├── integration/     # 整合測試
    └── unit/            # 單元測試

frontend/
├── src/
│   ├── components/      # UI 元件
│   ├── pages/           # 頁面 (域名列表、設定等)
│   ├── services/        # API 客戶端
│   └── main.js/tsx      # 應用程式入口
└── tests/
    └── e2e/             # 端對端測試

helm/
└── domain-manager/
    ├── Chart.yaml
    ├── values.yaml
    ├── templates/
    │   ├── deployment.yaml
    │   ├── service.yaml
    │   ├── ingress.yaml
    │   ├── pvc.yaml
    │   └── configmap.yaml
    └── README.md
```

**結構決策**: 選擇 Web 應用結構(後端 + 前端),因為需要提供 Web 介面、REST API 和 MCP 伺服器。後端處理所有 Kubernetes 互動和 SSL 管理邏輯,前端提供使用者介面。Helm Chart 確保簡單部署。

## 複雜度追蹤

> **僅在憲法檢查有需要說明的違規時填寫**

無違規需要說明。專案設計符合「簡單優先」原則。
