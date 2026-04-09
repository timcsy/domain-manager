# Quickstart: 多 Ingress Controller 支援

## 本地開發測試

```bash
cd backend
export K8S_MOCK=true
go run src/main.go
```

## 驗證步驟

### 1. 確認預設 Ingress Class

```bash
curl -H "X-Session-Token: <token>" http://localhost:8080/api/v1/settings | jq '.data[] | select(.key == "default_ingress_class")'
```

預期回傳 `"nginx"`。

### 2. 切換為 Traefik

```bash
curl -X PATCH http://localhost:8080/api/v1/settings \
  -H "X-Session-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{"settings": {"default_ingress_class": "traefik"}}'
```

### 3. 建立域名並驗證

```bash
curl -X POST http://localhost:8080/api/v1/domains \
  -H "X-Session-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{"domain_name": "test.example.com", "target_service": "my-app", "target_port": 80}'
```

在 Mock 模式下，檢查應用程式日誌確認 Ingress 使用 `traefik` class。

### K3s 部署

```bash
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  --set subdomains.defaultIngressClass=traefik \
  --set admin.password="your-password"
```
