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
- 🚧 MCP (Model Context Protocol) 支援 (開發中)

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
│   │   ├── middleware/  # HTTP 中介軟體
│   │   ├── k8s/         # Kubernetes 客戶端
│   │   └── db/          # 資料庫管理
│   ├── database/        # SQL 遷移腳本
│   └── Dockerfile       # 容器映像建置
├── frontend/         # Web 介面
│   └── src/
│       ├── pages/       # HTML 頁面
│       ├── components/  # UI 元件
│       └── styles/      # TailwindCSS 樣式
├── helm/             # Helm Chart
│   └── domain-manager/
│       ├── templates/   # K8s 資源模板
│       ├── Chart.yaml
│       └── values.yaml
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

完整 API 文件請參閱 `specs/001-k8s-domain-manager/contracts/api-rest.yaml`

### 主要 API 端點

```
POST   /api/v1/auth/login          # 登入
POST   /api/v1/auth/logout         # 登出
GET    /api/v1/domains             # 列出域名
POST   /api/v1/domains             # 新增域名
GET    /api/v1/domains/{id}        # 取得域名詳情
PUT    /api/v1/domains/{id}        # 更新域名
DELETE /api/v1/domains/{id}        # 刪除域名
GET    /api/v1/diagnostics/health  # 健康檢查
```

## 配置

主要配置透過 Helm values.yaml:

```yaml
admin:
  password: "your-password"
  email: "admin@example.com"

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
