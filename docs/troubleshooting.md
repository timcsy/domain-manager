# 故障排除指南

## 常見問題

### 無法登入

**症狀**：登入頁面顯示「Invalid credentials」

**解決方案**：
1. 確認使用預設帳號 `admin` / `admin`（或 Helm 安裝時設定的密碼）
2. 如果忘記密碼，刪除資料庫檔案重新初始化：
   ```bash
   # K8s 環境
   kubectl exec -n domain-manager deploy/domain-manager -- rm /data/database.db
   kubectl rollout restart -n domain-manager deploy/domain-manager
   ```

### SSL 憑證申請失敗

**症狀**：域名狀態停在 `pending`，SSL 憑證未產生

**診斷步驟**：
1. 確認 DNS 已正確指向 Ingress Controller IP
   ```bash
   nslookup your-domain.com
   ```
2. 確認 Let's Encrypt 郵箱已設定（系統設定頁面）
3. 查看診斷記錄（Web UI 或 API）
4. 如使用 staging 環境，瀏覽器會顯示不信任的憑證（正常行為）

**常見原因**：
- DNS 尚未生效（等待 TTL）
- Let's Encrypt rate limit（同一域名每週最多 5 張憑證）
- 防火牆阻擋 80/443 port

### Ingress 未生成

**症狀**：域名已建立但 Ingress 資源不存在

**診斷步驟**：
```bash
# 檢查 Ingress 資源
kubectl get ingress -A

# 檢查 domain-manager 日誌
kubectl logs -n domain-manager deploy/domain-manager
```

**常見原因**：
- RBAC 權限不足（檢查 ServiceAccount 和 ClusterRole）
- 目標 namespace 不存在
- 目標服務不存在

### 資料庫鎖定

**症狀**：API 回應 `database is locked`

**解決方案**：
SQLite 在高併發下可能遇到鎖定問題。系統已配置 WAL 模式和 busy timeout 來緩解，但如果持續發生：
1. 確認只有一個 replica（SQLite 不支援多實例）
2. 重啟 Pod
   ```bash
   kubectl rollout restart -n domain-manager deploy/domain-manager
   ```

### MCP 連線失敗

**症狀**：AI 工具無法連接 MCP endpoint

**診斷步驟**：
1. 確認 `/mcp` endpoint 可存取
   ```bash
   curl -X POST http://localhost:8080/mcp \
     -H "Content-Type: application/json" \
     -d '{"jsonrpc":"2.0","id":1,"method":"ping"}'
   ```
2. 確認回應包含 `"result":{}`
3. MCP endpoint 不需認證（公開端點）

### API Key 無法使用

**症狀**：使用 API Key 時回傳 401 Unauthorized

**檢查事項**：
- 確認使用 `X-API-Key` header（不是 `Authorization`）
- 確認金鑰未過期或被停用
- 確認金鑰有對應的操作權限（read/write/delete）

## 日誌與診斷

### 查看應用程式日誌

```bash
# 即時日誌
kubectl logs -f -n domain-manager deploy/domain-manager

# 最近 100 行
kubectl logs --tail=100 -n domain-manager deploy/domain-manager
```

### 查看診斷記錄

透過 API：
```bash
curl -H "X-Session-Token: <token>" http://localhost:8080/api/v1/diagnostics/logs?limit=50
```

透過 MCP：
```bash
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_system_health","arguments":{}}}'
```

### 健康檢查

```bash
curl http://localhost:8080/health
```

## 備份與還原

### 手動備份

透過 Web UI「備份管理」頁面，或透過 API：
```bash
curl -X POST -H "X-Session-Token: <token>" http://localhost:8080/api/v1/backup
```

### 還原

1. 下載備份檔案
2. 停止應用程式
3. 替換 `/data/database.db`
4. 重啟應用程式

## 取得協助

如果以上方法無法解決問題：
1. 收集應用程式日誌和診斷記錄
2. 記錄重現步驟
3. 在 GitHub Issues 提交問題報告
