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

## 配置選項

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
