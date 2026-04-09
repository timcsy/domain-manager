# Kubernetes 域名管理器

一個簡單易用的 Kubernetes Ingress 和 SSL 憑證管理工具，搭配 Web UI 讓 DevOps 工程師快速將子網域對應到 K8s 服務。

## 功能特色

- ✅ Web 介面管理域名，下拉選單選擇 K8s 服務
- ✅ 自動透過 Let's Encrypt 取得 SSL 憑證
- ✅ Cloudflare DNS-01 + cert-manager wildcard 憑證（全免費）
- ✅ 多 Ingress Controller 支援（Nginx / Traefik / K3s）
- ✅ 子網域批次操作、樹狀檢視
- ✅ API Key 認證 + RESTful API
- ✅ MCP Server（讓 AI 工具操作域名管理）
- ✅ 管理員帳號管理（密碼修改、email 修改）
- ✅ 資料庫備份與還原
- ✅ 單一 Helm Chart 部署（含 cert-manager）

## 快速開始

### Kubernetes 部署

**前置需求**：
- Kubernetes 1.20+ 叢集（含 K3s）
- Helm 3.0+

cert-manager 會隨 Helm chart 自動安裝，不需要手動安裝。

**安裝**：

```bash
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  --set admin.password="your-secure-password" \
  --set ingress.certManager.issuerEmail="your-email@example.com"
```

**K3s（Traefik）環境**：

```bash
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  --set admin.password="your-secure-password" \
  --set ingress.className=traefik \
  --set subdomains.defaultIngressClass=traefik
```

**存取 Web UI**：

```bash
kubectl port-forward -n domain-manager svc/domain-manager 8080:8080
open http://localhost:8080
```

預設登入：帳號 `admin`，密碼為安裝時設定的值（預設 `change-this-password`）。登入後請到系統設定修改密碼。

### 本地開發（無需 K8s）

```bash
cd backend
export K8S_MOCK=true
go run src/main.go

# 訪問 http://localhost:8080
# 帳號: admin / admin
```

## 使用指南

### 管理域名

1. 登入後進入「域名管理」
2. 點擊「新增域名」
3. 從下拉選單選擇目標 K8s 服務和埠號
4. SSL 模式選「自動」，系統自動申請 Let's Encrypt 憑證
5. 將域名 DNS A 記錄指向 Ingress Controller 的 LoadBalancer IP

### Cloudflare Wildcard 憑證

透過 Cloudflare 免費 DNS + DNS-01 challenge 申請 wildcard 憑證（`*.example.com`），不需要公開 80 port。

**設定方式 A：Web UI**

1. 在 [Cloudflare Dashboard](https://dash.cloudflare.com) 建立 API Token（權限：Zone - DNS - Edit）
2. 進入「系統設定」>「Cloudflare DNS-01 整合」，輸入 token
3. 建立域名時輸入 `*.example.com` 即可

**設定方式 B：Helm 部署時**

```bash
helm install domain-manager ./helm/domain-manager \
  --set cloudflare.enabled=true \
  --set cloudflare.apiToken="your-cloudflare-api-token"
```

### 管理員帳號

在「系統設定」>「帳號管理」可以：
- 修改管理員密碼（需驗證舊密碼，修改後自動登出）
- 修改管理員 email

### API Key

在「API 金鑰」頁面建立 API Key，用於程式化存取：

```bash
curl -H "X-API-Key: dm_your_key_here" http://localhost:8080/api/v1/domains
```

## 配置

### Helm values.yaml

```yaml
# 管理員帳號
admin:
  password: "your-password"
  email: "admin@example.com"

# Ingress Controller（nginx 或 traefik）
ingress:
  className: "nginx"
  certManager:
    enabled: true
    issuerEmail: "admin@example.com"
    acmeServer: "https://acme-v02.api.letsencrypt.org/directory"

# 子網域預設配置
subdomains:
  defaultSslMode: "auto"
  defaultIngressClass: "nginx"    # K3s 改為 "traefik"

# Cloudflare DNS-01（選填）
cloudflare:
  enabled: false
  apiToken: ""

# cert-manager（預設自動安裝）
cert-manager:
  enabled: true                   # 叢集已有 cert-manager 設為 false
  installCRDs: true

# 持久化儲存
persistence:
  enabled: true
  size: 1Gi

# 資源限制
resources:
  limits:
    cpu: 500m
    memory: 512Mi
```

### 環境變數

```bash
DATABASE_PATH="./data/database.db"
PORT="8080"
K8S_MOCK="false"                          # 開發模式
K8S_IN_CLUSTER="true"
LETSENCRYPT_EMAIL="admin@example.com"
LETSENCRYPT_STAGING="false"               # 測試用 staging 環境
DEFAULT_INGRESS_CLASS="nginx"             # 或 traefik
CLOUDFLARE_ENABLED="false"
CLOUDFLARE_API_TOKEN=""
FRONTEND_PATH="../frontend"
```

## API 文件

完整文件：[docs/api-usage.md](docs/api-usage.md) | MCP 範例：[docs/mcp-examples.md](docs/mcp-examples.md)

### 主要端點

```
POST   /api/v1/auth/login          # 登入
GET    /api/v1/domains             # 列出域名
POST   /api/v1/domains             # 新增域名
PUT    /api/v1/domains/{id}        # 更新域名
DELETE /api/v1/domains/{id}        # 刪除域名
GET    /api/v1/certificates        # 列出憑證
GET    /api/v1/services            # 列出 K8s 服務
GET    /api/v1/admin/profile       # 管理員資訊
PATCH  /api/v1/admin/password      # 修改密碼
GET    /api/v1/api-keys            # 列出 API 金鑰
POST   /api/v1/cloudflare/token    # 設定 Cloudflare Token
GET    /api/v1/cloudflare/status   # Cloudflare 狀態
POST   /api/v1/backup              # 建立備份
POST   /mcp                        # MCP JSON-RPC 2.0
```

## 開發

### 後端

```bash
cd backend
go mod download
go run src/main.go        # 開發伺服器
go build -o domain-manager src/main.go  # 建置
```

### 前端

```bash
cd frontend
npm install
npm run build:css         # 建置 CSS
npm run watch:css         # 監聽變更
```

### Docker

```bash
docker build -t domain-manager:latest .
```

映像發布在 `ghcr.io/timcsy/domain-manager`。

## 技術棧

| 層 | 技術 |
|---|------|
| 後端 | Go 1.22+, Chi, SQLite |
| 前端 | HTMX, TailwindCSS, Alpine.js |
| K8s | client-go, dynamic client (CRD) |
| SSL | cert-manager, lego (Let's Encrypt), Cloudflare DNS-01 |
| 部署 | Helm 3, Docker |

## 專案結構

```
domain-manager/
├── backend/          # Go 後端
│   ├── src/
│   │   ├── api/         # REST API handlers
│   │   ├── mcp/         # MCP Server
│   │   ├── services/    # 業務邏輯
│   │   ├── k8s/         # K8s 操作（Ingress, Secret, CRD）
│   │   └── ...
│   └── database/        # SQL migrations
├── frontend/         # HTMX + Alpine.js 前端
├── helm/             # Helm Chart（含 cert-manager 依賴）
├── docs/             # API、MCP、架構文件
└── specs/            # 功能規格
```

完整架構文件：[docs/architecture.md](docs/architecture.md)

## 授權

MIT License
