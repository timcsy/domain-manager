# Kubernetes 域名管理器 Helm Chart

這是 Kubernetes 域名管理器的官方 Helm Chart。

## 安裝

### 基本安裝

```bash
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace
```

### 自訂配置

建立 `values.yaml`:

```yaml
admin:
  password: "my-secure-password"
  email: "admin@mydomain.com"

ingress:
  className: "nginx"
  certManager:
    enabled: true
    issuerEmail: "admin@mydomain.com"

persistence:
  enabled: true
  size: 2Gi
  storageClass: "standard"

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi
```

然後安裝:

```bash
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  --values values.yaml
```

## 子域名管理

域名管理器支援統一管理多個子域名，自動配置 Ingress 和 SSL 憑證。

### 功能特點

- **自動化 Ingress 管理**：自動為每個子域名建立 Kubernetes Ingress 資源
- **SSL 憑證管理**：
  - 支援 Let's Encrypt 自動憑證申請與續期
  - 支援手動上傳自訂憑證
  - 支援不使用 SSL
- **健康檢查**：定期檢查子域名的 HTTP 可達性和回應時間
- **批次操作**：支援批次啟用、停用、刪除子域名

### 子域名配置選項

```yaml
subdomains:
  # 新子域名的預設 SSL 模式
  defaultSslMode: "auto"  # 可選: "none", "manual", "auto"

  # 預設 Ingress Class
  defaultIngressClass: "nginx"

  # 健康檢查設定
  healthCheck:
    enabled: true
    schedule: "*/5 * * * *"  # 每 5 分鐘檢查一次
    timeout: 10s
    healthyStatusCodes: [200, 201, 202, 203, 204, 205, 206, 300, 301, 302, 303, 304, 307, 308]
```

### 使用範例

透過 Web UI 或 API 添加子域名：

1. **Web UI**：登入後在「域名管理」頁面點擊「新增域名」
2. **API**：
```bash
curl -X POST http://domain-manager.example.com/api/v1/domains \
  -H "Content-Type: application/json" \
  -H "X-Session-Token: YOUR_TOKEN" \
  -d '{
    "domain_name": "app.example.com",
    "target_service": "my-app-service",
    "target_namespace": "default",
    "target_port": 8080,
    "ssl_mode": "auto"
  }'
```

添加後，系統會自動：
- 建立對應的 Ingress 資源
- 如果 `ssl_mode` 為 `auto`，自動申請 Let's Encrypt 憑證
- 開始定期健康檢查
- 在 Web UI 上顯示狀態

更多範例請參考 `examples/` 目錄。

## 配置選項

### 基本配置

| 參數 | 描述 | 預設值 |
|------|------|--------|
| `admin.password` | 管理員密碼 | `change-this-password` |
| `admin.email` | 管理員郵箱 | `admin@example.com` |
| `ingress.className` | Ingress Class 名稱 | `nginx` |
| `ingress.certManager.enabled` | 啟用 cert-manager | `true` |
| `persistence.enabled` | 啟用持久化儲存 | `true` |
| `persistence.size` | PVC 大小 | `1Gi` |
| `resources.limits.cpu` | CPU 限制 | `500m` |
| `resources.limits.memory` | 記憶體限制 | `512Mi` |

### 子域名配置

| 參數 | 描述 | 預設值 |
|------|------|--------|
| `subdomains.defaultSslMode` | 新子域名的預設 SSL 模式 | `auto` |
| `subdomains.defaultIngressClass` | 預設 Ingress Class | `nginx` |
| `subdomains.healthCheck.enabled` | 啟用健康檢查 | `true` |
| `subdomains.healthCheck.schedule` | 健康檢查排程 (cron) | `*/5 * * * *` |
| `subdomains.healthCheck.timeout` | 健康檢查超時時間 | `10s` |

## 升級

```bash
helm upgrade domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --values values.yaml
```

## 卸載

```bash
helm uninstall domain-manager --namespace domain-manager
```

## 故障排除

### 查看 Pod 狀態

```bash
kubectl get pods -n domain-manager
```

### 查看日誌

```bash
kubectl logs -n domain-manager deployment/domain-manager -f
```

### 查看事件

```bash
kubectl get events -n domain-manager --sort-by='.lastTimestamp'
```

## 最低需求

- Kubernetes 1.20+
- Helm 3.0+
- 至少 1 CPU 核心和 512MB 記憶體
