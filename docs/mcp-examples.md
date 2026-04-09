# MCP 客戶端使用範例

## Claude Desktop 設定

在 Claude Desktop 的設定檔中加入以下 MCP server 配置：

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "domain-manager": {
      "command": "curl",
      "args": ["-s", "-X", "POST", "http://localhost:8080/mcp", "-H", "Content-Type: application/json", "-d", "@-"]
    }
  }
}
```

> 注意：MCP endpoint 為 `POST /mcp`，使用 JSON-RPC 2.0 協議。

## 可用的 Tools

### 域名管理

| Tool | 說明 |
|------|------|
| `list_domains` | 列出所有域名，可依 status/service 篩選 |
| `get_domain` | 依 ID 取得域名詳情 |
| `create_domain` | 建立新的域名對應 |
| `update_domain` | 更新既有域名 |
| `delete_domain` | 刪除域名（支援軟/硬刪除） |

### 服務發現

| Tool | 說明 |
|------|------|
| `list_services` | 列出 Kubernetes 服務，可指定 namespace |

### 憑證管理

| Tool | 說明 |
|------|------|
| `get_certificate_status` | 取得特定憑證狀態 |
| `list_expiring_certificates` | 列出即將到期的憑證 |
| `renew_certificate` | 觸發憑證續期 |

### 診斷

| Tool | 說明 |
|------|------|
| `check_dns` | 檢查域名 DNS 解析 |
| `get_diagnostics` | 取得診斷記錄 |
| `get_system_health` | 取得系統整體健康狀態 |

## 可用的 Resources

| URI | 說明 |
|-----|------|
| `domain://list` | 所有域名列表 |
| `domain://{name}` | 特定域名詳情 |
| `service://list` | Kubernetes 服務列表 |
| `certificate://list` | SSL 憑證列表 |
| `diagnostics://logs` | 診斷記錄 |

## curl 使用範例

### 初始化連線

```bash
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {"name": "test-client", "version": "1.0"}
    }
  }'
```

### 列出所有工具

```bash
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}'
```

### 呼叫工具：列出域名

```bash
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
      "name": "list_domains",
      "arguments": {}
    }
  }'
```

### 呼叫工具：建立域名

```bash
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
      "name": "create_domain",
      "arguments": {
        "domain_name": "app.example.com",
        "target_service": "my-app",
        "target_namespace": "default",
        "target_port": 8080
      }
    }
  }'
```

### 讀取資源：域名列表

```bash
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "resources/read",
    "params": {"uri": "domain://list"}
  }'
```

### 取得系統健康狀態

```bash
curl -s -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
      "name": "get_system_health",
      "arguments": {}
    }
  }'
```
