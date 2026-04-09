# API 使用文件

## 認證

### Session Token（Web UI）

登入後取得 session token，透過 `X-Session-Token` header 傳遞：

```bash
# 登入
curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "admin"}'

# 使用 token
curl -H "X-Session-Token: <token>" http://localhost:8080/api/v1/domains
```

### API Key（程式化存取）

透過 Web UI 或 API 建立 API Key，使用 `X-API-Key` header：

```bash
curl -H "X-API-Key: dm_your_key_here" http://localhost:8080/api/v1/domains
```

## API 端點

### 域名管理

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/v1/domains` | 列出所有域名 |
| POST | `/api/v1/domains` | 建立域名 |
| GET | `/api/v1/domains/{id}` | 取得域名詳情 |
| PUT | `/api/v1/domains/{id}` | 更新域名 |
| DELETE | `/api/v1/domains/{id}` | 刪除域名 |
| GET | `/api/v1/domains/{id}/status` | 取得域名狀態 |
| GET | `/api/v1/domains/{id}/diagnostics` | 取得域名診斷 |
| GET | `/api/v1/domains/tree` | 取得域名樹狀結構 |
| POST | `/api/v1/domains/batch` | 批次建立域名 |
| PATCH | `/api/v1/domains/batch` | 批次更新域名 |
| DELETE | `/api/v1/domains/batch` | 批次刪除域名 |

### 憑證管理

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/v1/certificates` | 列出所有憑證 |
| POST | `/api/v1/certificates` | 上傳憑證 |
| GET | `/api/v1/certificates/expiring` | 列出即將到期憑證 |
| GET | `/api/v1/certificates/{id}` | 取得憑證詳情 |
| DELETE | `/api/v1/certificates/{id}` | 刪除憑證 |
| POST | `/api/v1/certificates/{id}/renew` | 續期憑證 |

### Kubernetes 服務

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/v1/services` | 列出所有 K8s 服務 |
| GET | `/api/v1/services/{namespace}/{name}` | 取得特定服務 |

### 診斷

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/v1/diagnostics/health` | 系統健康檢查 |
| GET | `/api/v1/diagnostics/logs` | 取得診斷記錄 |

### 系統設定

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/v1/settings` | 取得系統設定 |
| PATCH | `/api/v1/settings` | 更新系統設定 |

### 管理員帳號

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/v1/admin/profile` | 取得管理員基本資訊 |
| PATCH | `/api/v1/admin/password` | 修改密碼（需舊密碼驗證） |
| PATCH | `/api/v1/admin/email` | 修改 email |

### Cloudflare DNS-01

| 方法 | 路徑 | 說明 |
|------|------|------|
| POST | `/api/v1/cloudflare/token` | 設定 Cloudflare API Token（驗證後儲存） |
| GET | `/api/v1/cloudflare/status` | 取得 Cloudflare 整合狀態 |
| DELETE | `/api/v1/cloudflare/token` | 移除 Cloudflare Token（停用 DNS-01） |

### API 金鑰

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/v1/api-keys` | 列出所有 API 金鑰 |
| POST | `/api/v1/api-keys` | 建立 API 金鑰 |
| DELETE | `/api/v1/api-keys/{id}` | 刪除 API 金鑰 |

### 備份

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/api/v1/backup` | 列出所有備份 |
| POST | `/api/v1/backup` | 建立備份 |
| GET | `/api/v1/backup/{filename}` | 下載備份 |
| DELETE | `/api/v1/backup/{filename}` | 刪除備份 |

### MCP

| 方法 | 路徑 | 說明 |
|------|------|------|
| POST | `/mcp` | MCP JSON-RPC 2.0 端點 |

### 其他

| 方法 | 路徑 | 說明 |
|------|------|------|
| GET | `/health` | 服務健康檢查（公開） |

## 回應格式

### 成功

```json
{
  "success": true,
  "data": { ... },
  "message": "Operation successful"
}
```

### 錯誤

```json
{
  "success": false,
  "error": "Bad Request",
  "message": "Detailed error description"
}
```

### 分頁

```json
{
  "success": true,
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 50,
    "total_pages": 3
  }
}
```

## 範例

### 建立域名

```bash
curl -X POST http://localhost:8080/api/v1/domains \
  -H "X-API-Key: dm_your_key" \
  -H "Content-Type: application/json" \
  -d '{
    "domain_name": "app.example.com",
    "target_service": "my-app",
    "target_namespace": "default",
    "target_port": 8080,
    "ssl_mode": "auto"
  }'
```

### 建立 API 金鑰

```bash
curl -X POST http://localhost:8080/api/v1/api-keys \
  -H "X-Session-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "key_name": "CI/CD Pipeline",
    "permissions": ["read", "write"]
  }'
```

### 列出即將到期的憑證

```bash
curl -H "X-API-Key: dm_your_key" \
  "http://localhost:8080/api/v1/certificates/expiring?days=30"
```

### 設定 Cloudflare API Token

```bash
curl -X POST http://localhost:8080/api/v1/cloudflare/token \
  -H "X-Session-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{"api_token": "your-cloudflare-api-token"}'
```

### 查看 Cloudflare 整合狀態

```bash
curl -H "X-Session-Token: <token>" \
  http://localhost:8080/api/v1/cloudflare/status
```

### 建立 Wildcard 域名

```bash
curl -X POST http://localhost:8080/api/v1/domains \
  -H "X-Session-Token: <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "domain_name": "*.example.com",
    "target_service": "my-app",
    "target_port": 80,
    "ssl_mode": "auto"
  }'
```
