# 系統架構

## 概覽

```
┌─────────────────────────────────────────────────┐
│                    使用者                         │
│   Web UI    CLI/curl    AI 工具 (Claude Desktop) │
└──────┬──────────┬──────────────┬────────────────┘
       │          │              │
       ▼          ▼              ▼
┌──────────────────────────────────────────────┐
│              HTTP Server (Chi)                │
│  ┌─────────────────────────────────────────┐ │
│  │ Middleware: Auth, RateLimit, Tracing,    │ │
│  │            CORS, Logger, Recoverer      │ │
│  └─────────────────────────────────────────┘ │
│  ┌──────────────┐  ┌───────────────────────┐ │
│  │  REST API     │  │    MCP Server         │ │
│  │  /api/v1/*    │  │    /mcp               │ │
│  │               │  │    JSON-RPC 2.0       │ │
│  └──────┬───────┘  └──────────┬────────────┘ │
└─────────┼──────────────────────┼─────────────┘
          │                      │
          ▼                      ▼
┌──────────────────────────────────────────────┐
│              Service Layer                    │
│  ┌────────────┐ ┌──────────┐ ┌────────────┐ │
│  │ Domain     │ │ Cert     │ │ APIKey     │ │
│  │ Service    │ │ Service  │ │ Service    │ │
│  └──────┬─────┘ └────┬─────┘ └─────┬──────┘ │
│  ┌──────┴─────┐ ┌────┴─────┐ ┌─────┴──────┐ │
│  │ Auth       │ │ Settings │ │ Backup     │ │
│  │ Service    │ │ Service  │ │ Service    │ │
│  └────────────┘ └──────────┘ └────────────┘ │
└──────────┬──────────────────────┬────────────┘
           │                      │
     ┌─────▼──────┐        ┌─────▼──────┐
     │  SQLite    │        │ Kubernetes │
     │  (WAL)     │        │ client-go  │
     └────────────┘        └─────┬──────┘
                                 │
                           ┌─────▼──────┐
                           │ K8s API    │
                           │ Server     │
                           └────────────┘
```

## 元件說明

### HTTP Server

- **框架**：go-chi/chi v5
- **埠號**：8080（可配置）
- **中介軟體**：RequestTracing → RateLimit → CORS → RequestID → RealIP → Logger → Recoverer → Compress

### REST API (`/api/v1/`)

處理所有 CRUD 操作，支援兩種認證方式：
- **Session Token**：Web UI 登入後取得，透過 `X-Session-Token` header
- **API Key**：程式化存取，透過 `X-API-Key` header，SHA-256 hash 儲存

### MCP Server (`/mcp`)

實作 MCP (Model Context Protocol)，使用 JSON-RPC 2.0：
- **Tools**：list_domains, create_domain, list_services, get_system_health 等
- **Resources**：domain://list, service://list, certificate://list 等
- 無需認證（設計為叢集內部使用）

### Service Layer

業務邏輯層，負責：
- 域名 CRUD + Kubernetes Ingress 同步
- SSL 憑證管理 + Let's Encrypt 自動化（透過 lego 庫）
- 子網域驗證與衝突檢測
- API Key 產生與驗證
- 資料庫備份

### 資料層

- **SQLite**：WAL 模式，單一檔案資料庫
- **表格**：domains, certificates, diagnostic_logs, admin_accounts, api_keys, system_settings
- **限制**：單一實例（SQLite 不支援多 writer）

### Kubernetes 整合

- **client-go**：直接操作 Ingress, Service, Secret, Namespace 資源
- **ServiceManager**：列出和查詢 K8s 服務（支援 Mock 模式）
- **RBAC**：需要 ClusterRole 權限操作 Ingress 和 Secret

## 資料流

### 建立域名

```
使用者 → POST /api/v1/domains
  → Auth middleware 驗證
  → DomainService.CreateDomain()
    → DomainRepository.Create() (SQLite)
    → [非同步] K8s: 建立 Ingress 資源
    → [非同步] CertService: 申請 Let's Encrypt 憑證
      → lego: ACME challenge
      → K8s: 建立 TLS Secret
      → 更新 Ingress TLS 設定
```

### MCP 工具呼叫

```
AI 工具 → POST /mcp {"method":"tools/call","params":{"name":"list_domains"}}
  → MCP Server.HandleMessage()
  → 解析 JSON-RPC 2.0
  → 路由到 handleListDomains()
  → DomainService.ListDomains()
  → 回傳 JSON-RPC 2.0 response
```

## 部署架構

```
┌─ Kubernetes Cluster ─────────────────────────┐
│                                               │
│  ┌─ Namespace: domain-manager ─────────────┐ │
│  │                                          │ │
│  │  ┌──────────────────────────┐           │ │
│  │  │ Deployment               │           │ │
│  │  │ domain-manager           │           │ │
│  │  │  - Go backend            │           │ │
│  │  │  - HTMX frontend        │           │ │
│  │  │  - SQLite (PVC)          │           │ │
│  │  └──────────┬───────────────┘           │ │
│  │             │                            │ │
│  │  ┌──────────▼───────────────┐           │ │
│  │  │ Service (ClusterIP:8080) │           │ │
│  │  └──────────────────────────┘           │ │
│  │                                          │ │
│  │  ┌──────────────────────────┐           │ │
│  │  │ PersistentVolumeClaim    │           │ │
│  │  │ /data (1Gi)              │           │ │
│  │  └──────────────────────────┘           │ │
│  │                                          │ │
│  │  ┌──────────────────────────┐           │ │
│  │  │ ServiceAccount + RBAC    │           │ │
│  │  │ (ClusterRole)            │           │ │
│  │  └──────────────────────────┘           │ │
│  └──────────────────────────────────────────┘ │
│                                               │
│  ┌─ Ingress Controller (nginx) ────────────┐ │
│  │  管理由 domain-manager 建立的 Ingress     │ │
│  └──────────────────────────────────────────┘ │
│                                               │
│  ┌─ cert-manager (選擇性) ─────────────────┐ │
│  │  ClusterIssuer: Let's Encrypt            │ │
│  └──────────────────────────────────────────┘ │
└───────────────────────────────────────────────┘
```

## 技術選型理由

| 選擇 | 理由 |
|------|------|
| **Go** | 單一二進位檔部署，原生 K8s client-go 支援 |
| **SQLite** | 零依賴，單一檔案，適合單實例部署 |
| **HTMX + Alpine.js** | 輕量前端，無需複雜建置流程 |
| **Chi** | 輕量 HTTP router，與 net/http 完全相容 |
| **lego** | Go 原生 ACME client，不依賴 cert-manager |
