# Data Model: Cloudflare DNS + cert-manager DNS-01 整合

## 既有實體（擴展）

### system_settings

新增設定項：

| Key | 類型 | 預設值 | 說明 |
|-----|------|--------|------|
| `cloudflare_api_token` | string | `""` | Cloudflare API Token（加密儲存） |
| `cloudflare_enabled` | string | `"0"` | 是否啟用 Cloudflare DNS-01 整合 |

## 新增 K8s 資源（非資料庫）

### Secret: cloudflare-api-token

由應用程式動態建立在 cert-manager namespace。

| 欄位 | 值 |
|------|-----|
| name | `cloudflare-api-token` |
| namespace | `cert-manager`（或 cert-manager 安裝的 namespace） |
| type | `Opaque` |
| data.api-token | Cloudflare API Token |

### ClusterIssuer: letsencrypt-prod（修改既有）

在既有 ClusterIssuer 中加入 DNS-01 solver：

| Solver | 用途 | Selector |
|--------|------|----------|
| dns01 (cloudflare) | wildcard 域名 | `dnsNames: ["*.example.com"]` |
| http01 (ingress) | 一般域名 | 無 selector（fallback） |

## 程式碼層面實體

### CloudflareService

| 方法 | 說明 |
|------|------|
| ValidateToken(token) | 呼叫 Cloudflare API 驗證 token 有效性 |
| SaveToken(token) | 儲存 token 到 system_settings + 建立 K8s Secret |
| GetTokenStatus() | 回傳 token 是否已設定且有效 |

### CertManagerHelper (k8s/certmanager.go)

| 方法 | 說明 |
|------|------|
| CreateOrUpdateCloudflareSecret(token) | 建立/更新 Cloudflare token Secret |
| CreateOrUpdateClusterIssuer(config) | 建立/更新含 DNS-01 solver 的 ClusterIssuer |
| GetClusterIssuerStatus() | 取得 ClusterIssuer 狀態 |

## 狀態流程

```
使用者輸入 Token
  → ValidateToken（Cloudflare API）
  → SaveToken → system_settings + K8s Secret
  → CreateOrUpdateClusterIssuer（加入 DNS-01 solver）
  → 建立 wildcard 域名
  → cert-manager 自動申請憑證（DNS-01 challenge）
  → TLS Secret 建立
  → Ingress 使用 TLS Secret
```
