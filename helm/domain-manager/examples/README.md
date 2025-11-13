# Helm 配置範例

此目錄包含 Domain Manager 的 Helm 配置範例檔案。

## 範例檔案

### 1. basic-subdomains.yaml

基本子域名配置範例，適合開發和測試環境。

**特點**:
- 啟用 Let's Encrypt 自動 SSL
- 使用 nginx ingress controller
- 每 5 分鐘健康檢查
- 基本資源配置

**使用方式**:
```bash
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  -f examples/basic-subdomains.yaml
```

### 2. production.yaml

生產環境配置範例，包含高可用性和安全性最佳實踐。

**特點**:
- 2 個副本提供高可用性
- Pod 反親和性確保副本分佈
- 每 2 分鐘健康檢查
- 每日自動備份
- 更高的資源限制
- 節點選擇和容忍度配置

**使用方式**:
```bash
# 建議先建立 Secret 存放敏感資訊
kubectl create secret generic admin-credentials \
  --from-literal=password='YOUR_SECURE_PASSWORD' \
  -n domain-manager

helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  -f examples/production.yaml
```

### 3. multiple-ingress-classes.yaml

多 Ingress Controller 配置範例。

**特點**:
- Web UI 使用 nginx ingress controller
- 子域名預設使用 traefik ingress controller
- 適合集群中有多種 Ingress Controller 的場景

**使用範例**:
```bash
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  -f examples/multiple-ingress-classes.yaml
```

**使用場景**:
- 內部服務使用 nginx，外部 API 使用 traefik
- 根據流量特性選擇不同的 Ingress Controller
- A/B 測試不同的 Ingress Controller

## 自訂配置

您可以組合多個配置檔案：

```bash
helm install domain-manager ./helm/domain-manager \
  -f examples/basic-subdomains.yaml \
  -f my-custom-values.yaml
```

後面的檔案會覆蓋前面的配置。

## 配置驗證

安裝後驗證配置：

```bash
# 查看 Pod 狀態
kubectl get pods -n domain-manager

# 查看日誌
kubectl logs -n domain-manager deployment/domain-manager -f

# 查看 Ingress
kubectl get ingress -n domain-manager

# 查看持久化卷
kubectl get pvc -n domain-manager
```

## 更新配置

如需更新配置：

```bash
helm upgrade domain-manager ./helm/domain-manager \
  -f examples/production.yaml \
  --namespace domain-manager
```

## 安全建議

1. **密碼管理**: 不要在配置檔案中使用明文密碼，使用 Kubernetes Secrets
2. **SSL 憑證**: 生產環境務必使用正式的 Let's Encrypt 憑證（非 staging）
3. **資源限制**: 根據實際負載調整資源配置
4. **備份**: 啟用並定期測試備份恢復
5. **監控**: 整合 Prometheus 等監控工具

## 故障排除

如果遇到問題，請參考主 README.md 的故障排除章節，或執行：

```bash
helm test domain-manager -n domain-manager
```
