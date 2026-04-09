# Quickstart: Cloudflare DNS + cert-manager DNS-01

## 前置需求

1. Cloudflare 帳號（免費方案即可）
2. 已在 Cloudflare 設定 DNS zone
3. Cloudflare API Token（權限：Zone - DNS - Edit）

## 建立 Cloudflare API Token

1. 登入 [Cloudflare Dashboard](https://dash.cloudflare.com)
2. 進入 My Profile → API Tokens
3. 點擊「Create Token」
4. 選擇「Edit zone DNS」template
5. 設定 Zone Resources：選擇你的 zone
6. 建立並複製 token

## 方式 A：透過 UI 設定

1. 登入 domain-manager Web UI
2. 進入「系統設定」頁面
3. 在 Cloudflare 區塊輸入 API Token
4. 點擊「驗證並儲存」
5. 確認狀態顯示為「已啟用」

## 方式 B：透過 Helm 預設

```bash
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  --set admin.password="your-password" \
  --set cloudflare.apiToken="your-cloudflare-token" \
  --set cloudflare.enabled=true
```

## 驗證

### 檢查 ClusterIssuer

```bash
kubectl get clusterissuer letsencrypt-prod -o yaml
# 應包含 dns01 solver with cloudflare
```

### 申請 Wildcard 憑證

1. 在 UI 建立域名，SSL 模式選「wildcard」
2. 或建立 `*.yourdomain.com` 的域名
3. 檢查 cert-manager 日誌：
   ```bash
   kubectl logs -n cert-manager deploy/cert-manager -f
   ```
4. 確認 Certificate 狀態：
   ```bash
   kubectl get certificate -A
   ```

### 檢查 Cloudflare 狀態 API

```bash
curl -H "X-Session-Token: <token>" http://localhost:8080/api/v1/cloudflare/status
```
