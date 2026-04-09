# Kubernetes 域名管理器

一個簡單易用的 Kubernetes Ingress 和 SSL 憑證管理工具。

## 功能特色

- ✅ 簡單的 Web 介面管理域名
- ✅ 自動透過 Let's Encrypt 取得 SSL 憑證
- ✅ 支援自訂憑證上傳
- ✅ Kubernetes Ingress 自動配置
- ✅ 單一 Helm Chart 部署
- ✅ 輕量級設計 (≤512MB 記憶體)
- ✅ RESTful API 支援
- ✅ API Key 認證機制
- ✅ MCP (Model Context Protocol) 支援
- ✅ 資料庫備份與還原
- ✅ 速率限制與請求追蹤
- ✅ 系統設定管理介面
- ✅ Cloudflare DNS-01 + cert-manager wildcard 憑證
- ✅ 多 Ingress Controller 支援（Nginx / Traefik）

## 快速開始

### 🔧 本地開發 (無需 K8s)

```bash
cd backend
cp .env.local .env  # 設定 K8S_MOCK=true
go run src/main.go

# 訪問 http://localhost:8080
# 帳號: admin / admin
```

### ☸️ Kubernetes 部署

**前置需求**:
- Kubernetes 1.20+ 叢集
- Helm 3.0+
- Nginx Ingress Controller (推薦)
- cert-manager (可選,用於自動 SSL)

### 安裝

1. 新增 Helm repository:

```bash
helm repo add domain-manager https://your-repo-url.com/charts
helm repo update
```

2. 安裝 domain-manager:

```bash
helm install domain-manager domain-manager/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  --set admin.password="your-secure-password"
```

3. 存取 Web 介面:

```bash
# Port forward 到本地
kubectl port-forward -n domain-manager svc/domain-manager 8080:8080

# 在瀏覽器開啟
open http://localhost:8080
```

預設登入資訊:
- 使用者名稱: `admin`
- 密碼: `admin` (請立即修改!)

## 專案結構

```
domain-manager/
├── backend/          # Go 後端服務
│   ├── src/
│   │   ├── models/      # 資料模型
│   │   ├── repositories/# 資料存取層
│   │   ├── services/    # 業務邏輯
│   │   ├── api/         # REST API
│   │   ├── mcp/         # MCP 伺服器
│   │   ├── middleware/  # HTTP 中介軟體
│   │   ├── k8s/         # Kubernetes 客戶端
│   │   └── db/          # 資料庫管理
│   ├── database/        # SQL 遷移腳本
│   └── tests/           # 測試腳本
├── frontend/         # Web 介面
│   └── src/
│       ├── pages/       # HTML 頁面
│       ├── components/  # UI 元件
│       └── styles/      # TailwindCSS 樣式
├── helm/             # Helm Chart
│   └── domain-manager/
├── docs/             # 文件
│   ├── api-usage.md     # API 使用文件
│   ├── mcp-examples.md  # MCP 範例
│   └── postman/         # Postman collection
└── specs/            # 功能規格文件
```

## 開發

### 後端開發

```bash
cd backend

# 安裝依賴
go mod download

# 運行開發伺服器
go run src/main.go

# 建置
go build -o domain-manager src/main.go
```

### 前端開發

```bash
cd frontend

# 安裝依賴
npm install

# 建置 CSS
npm run build:css

# 監聽 CSS 變更
npm run watch:css
```

### 建置 Docker 映像

```bash
cd backend
docker build -t domain-manager:latest .
```

## 技術棧

- **後端**: Go 1.22+, Chi, SQLite
- **前端**: HTMX, TailwindCSS, Alpine.js
- **Kubernetes**: client-go, Ingress
- **SSL**: cert-manager, lego (Let's Encrypt)
- **部署**: Helm 3

## API 文件

完整 API 文件請參閱 [docs/api-usage.md](docs/api-usage.md)。MCP 使用範例請參閱 [docs/mcp-examples.md](docs/mcp-examples.md)。

### 主要 API 端點

```
POST   /api/v1/auth/login          # 登入
POST   /api/v1/auth/logout         # 登出
GET    /api/v1/domains             # 列出域名
POST   /api/v1/domains             # 新增域名
PUT    /api/v1/domains/{id}        # 更新域名
DELETE /api/v1/domains/{id}        # 刪除域名
GET    /api/v1/certificates        # 列出憑證
GET    /api/v1/services            # 列出 K8s 服務
GET    /api/v1/api-keys            # 列出 API 金鑰
POST   /api/v1/backup              # 建立備份
POST   /api/v1/cloudflare/token    # 設定 Cloudflare Token
GET    /api/v1/cloudflare/status   # Cloudflare 整合狀態
POST   /mcp                        # MCP JSON-RPC 2.0 端點
```

## Cloudflare DNS-01 Wildcard 憑證

透過 Cloudflare 免費 DNS + cert-manager DNS-01 solver 自動申請 wildcard 憑證，全免費方案。

### 設定方式

**方式 A：Web UI**

1. 在 [Cloudflare Dashboard](https://dash.cloudflare.com) 建立 API Token（權限：Zone - DNS - Edit）
2. 進入系統設定頁面，在 Cloudflare 區塊輸入 token
3. 建立域名時輸入 `*.example.com` 即可申請 wildcard 憑證

**方式 B：Helm 部署時設定**

```bash
helm install domain-manager ./helm/domain-manager \
  --set cloudflare.enabled=true \
  --set cloudflare.apiToken="your-cloudflare-api-token"
```

## 配置

### 環境變數

後端支援以下環境變數配置:

```bash
# 資料庫設定
DATABASE_PATH="./data/database.db"

# Kubernetes 設定
K8S_MOCK="false"                    # 開發模式 (不連接實際 K8s 叢集)
K8S_IN_CLUSTER="true"               # 是否在 K8s 叢集內運行

# Let's Encrypt 設定
LETSENCRYPT_EMAIL="admin@example.com"           # 聯絡信箱 (用於憑證到期通知)
LETSENCRYPT_ACCOUNT_PATH="./data/letsencrypt"  # 帳戶資料儲存路徑
LETSENCRYPT_STAGING="false"                     # 使用 staging 環境 (測試用)

# 伺服器設定
PORT="8080"                         # HTTP 服務埠號
FRONTEND_PATH="../frontend"         # 前端檔案路徑
```

**重要**: Let's Encrypt staging 環境用於測試，不會簽發真實憑證。生產環境請設定 `LETSENCRYPT_STAGING=false`。

### Helm values.yaml 配置

主要配置透過 Helm values.yaml:

```yaml
admin:
  password: "your-password"
  email: "admin@example.com"

# Let's Encrypt 配置
letsencrypt:
  email: "admin@example.com"        # 用於憑證到期通知
  staging: false                    # 生產環境使用 false，測試使用 true

ingress:
  className: "nginx"
  certManager:
    enabled: true
    issuerEmail: "admin@example.com"

persistence:
  enabled: true
  size: 1Gi

resources:
  limits:
    cpu: 500m
    memory: 512Mi
```

## 授權

MIT License

## 貢獻

歡迎提交 Issue 和 Pull Request!

## 支援

- GitHub Issues: https://github.com/your-repo/domain-manager/issues
- 文件: https://docs.example.com
