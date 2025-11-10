# MCP 協議規格: Kubernetes 域名管理器

**功能分支**: `001-k8s-domain-manager`
**建立日期**: 2025-11-07
**狀態**: Draft
**版本**: 1.0
**協議版本**: MCP 0.1.0

---

## 1. 概述

本文件定義 Kubernetes 域名管理器的 MCP (Model Context Protocol) 伺服器實作規格,使 AI 助理(如 Claude Desktop)可以透過標準化協議管理域名配置。

### 1.1 協議基礎

- **傳輸層**: HTTP/SSE (Server-Sent Events)
- **訊息格式**: JSON-RPC 2.0
- **認證方式**: API Key (透過 `X-API-Key` HTTP header)
- **端點**: `/mcp` (相對於 API base URL)

### 1.2 連接資訊

```json
{
  "name": "k8s-domain-manager",
  "version": "1.0.0",
  "protocol_version": "0.1.0",
  "capabilities": {
    "tools": true,
    "resources": true,
    "prompts": false
  }
}
```

---

## 2. Tools (工具列表)

MCP Tools 讓 AI 助理可以執行具體操作。以下定義所有可用工具。

### 2.1 list_domains

列出所有域名配置。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "status": {
      "type": "string",
      "enum": ["pending", "active", "error", "deleted"],
      "description": "篩選特定狀態的域名"
    },
    "enabled": {
      "type": "boolean",
      "description": "篩選啟用或停用的域名"
    },
    "limit": {
      "type": "integer",
      "default": 20,
      "minimum": 1,
      "maximum": 100,
      "description": "限制返回的域名數量"
    }
  }
}
```

**Output**:
```json
{
  "domains": [
    {
      "id": 1,
      "domain_name": "example.com",
      "target_service": "nginx-service",
      "target_namespace": "default",
      "target_port": 80,
      "ssl_mode": "auto",
      "status": "active",
      "enabled": true,
      "created_at": "2025-11-07T10:00:00Z",
      "updated_at": "2025-11-07T10:00:00Z"
    }
  ],
  "total": 1
}
```

**範例請求** (JSON-RPC 2.0):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "list_domains",
    "arguments": {
      "status": "active",
      "limit": 10
    }
  }
}
```

---

### 2.2 get_domain

取得單一域名的詳細資訊。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "domain_name": {
      "type": "string",
      "description": "域名名稱 (例如 example.com)"
    },
    "id": {
      "type": "integer",
      "description": "域名 ID"
    }
  },
  "oneOf": [
    {"required": ["domain_name"]},
    {"required": ["id"]}
  ]
}
```

**Output**:
```json
{
  "id": 1,
  "domain_name": "example.com",
  "target_service": "nginx-service",
  "target_namespace": "default",
  "target_port": 80,
  "ssl_mode": "auto",
  "certificate_id": 1,
  "status": "active",
  "enabled": true,
  "created_at": "2025-11-07T10:00:00Z",
  "updated_at": "2025-11-07T10:00:00Z",
  "certificate": {
    "id": 1,
    "valid_until": "2026-02-05T10:00:00Z",
    "status": "valid",
    "source": "letsencrypt"
  }
}
```

---

### 2.3 create_domain

建立新的域名配置。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "domain_name": {
      "type": "string",
      "description": "域名名稱 (例如 api.example.com)",
      "pattern": "^([a-zA-Z0-9]([a-zA-Z0-9\\-]{0,61}[a-zA-Z0-9])?\\.)+[a-zA-Z]{2,}$"
    },
    "target_service": {
      "type": "string",
      "description": "目標 Kubernetes Service 名稱"
    },
    "target_namespace": {
      "type": "string",
      "default": "default",
      "description": "目標 Service 所在的命名空間"
    },
    "target_port": {
      "type": "integer",
      "minimum": 1,
      "maximum": 65535,
      "description": "目標 Service 的連接埠"
    },
    "ssl_mode": {
      "type": "string",
      "enum": ["auto", "manual"],
      "default": "auto",
      "description": "SSL 憑證模式 (auto: Let's Encrypt, manual: 使用者上傳)"
    }
  },
  "required": ["domain_name", "target_service", "target_port"]
}
```

**Output**:
```json
{
  "id": 2,
  "domain_name": "api.example.com",
  "status": "pending",
  "message": "域名配置已建立,正在申請 SSL 憑證..."
}
```

---

### 2.4 update_domain

更新現有域名配置。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "integer",
      "description": "域名 ID"
    },
    "domain_name": {
      "type": "string",
      "description": "域名名稱 (用於查詢)"
    },
    "target_service": {
      "type": "string",
      "description": "新的目標 Service 名稱"
    },
    "target_namespace": {
      "type": "string",
      "description": "新的命名空間"
    },
    "target_port": {
      "type": "integer",
      "minimum": 1,
      "maximum": 65535,
      "description": "新的連接埠"
    },
    "enabled": {
      "type": "boolean",
      "description": "啟用或停用域名"
    }
  },
  "oneOf": [
    {"required": ["id"]},
    {"required": ["domain_name"]}
  ]
}
```

**Output**:
```json
{
  "id": 1,
  "domain_name": "example.com",
  "status": "active",
  "message": "域名配置已更新"
}
```

---

### 2.5 delete_domain

刪除域名配置。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "id": {
      "type": "integer",
      "description": "域名 ID"
    },
    "domain_name": {
      "type": "string",
      "description": "域名名稱"
    },
    "hard": {
      "type": "boolean",
      "default": false,
      "description": "是否永久刪除 (預設為軟刪除)"
    }
  },
  "oneOf": [
    {"required": ["id"]},
    {"required": ["domain_name"]}
  ]
}
```

**Output**:
```json
{
  "success": true,
  "message": "域名已刪除"
}
```

---

### 2.6 list_services

列出所有可用的 Kubernetes Services。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "namespace": {
      "type": "string",
      "description": "篩選特定命名空間的服務"
    },
    "label_selector": {
      "type": "string",
      "description": "使用標籤篩選服務 (例如 app=nginx)"
    }
  }
}
```

**Output**:
```json
{
  "services": [
    {
      "name": "nginx-service",
      "namespace": "default",
      "type": "ClusterIP",
      "cluster_ip": "10.96.0.10",
      "ports": [
        {
          "name": "http",
          "port": 80,
          "target_port": 8080,
          "protocol": "TCP"
        }
      ],
      "labels": {
        "app": "nginx",
        "version": "v1"
      }
    }
  ],
  "total": 1
}
```

---

### 2.7 get_certificate_status

查詢 SSL 憑證狀態。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "domain_name": {
      "type": "string",
      "description": "域名名稱"
    },
    "certificate_id": {
      "type": "integer",
      "description": "憑證 ID"
    }
  },
  "oneOf": [
    {"required": ["domain_name"]},
    {"required": ["certificate_id"]}
  ]
}
```

**Output**:
```json
{
  "certificate_id": 1,
  "domain_name": "example.com",
  "source": "letsencrypt",
  "issuer": "Let's Encrypt Authority X3",
  "valid_from": "2025-11-07T10:00:00Z",
  "valid_until": "2026-02-05T10:00:00Z",
  "status": "valid",
  "days_until_expiry": 90,
  "auto_renew": true,
  "last_renewal_attempt": null,
  "renewal_error": null
}
```

---

### 2.8 list_expiring_certificates

列出即將到期的憑證。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "days": {
      "type": "integer",
      "default": 30,
      "minimum": 1,
      "maximum": 90,
      "description": "查詢幾天內到期的憑證"
    }
  }
}
```

**Output**:
```json
{
  "certificates": [
    {
      "id": 2,
      "domain_name": "api.example.com",
      "valid_until": "2025-12-07T10:00:00Z",
      "days_until_expiry": 30,
      "status": "expiring",
      "auto_renew": true
    }
  ],
  "total": 1
}
```

---

### 2.9 renew_certificate

手動觸發憑證續約。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "certificate_id": {
      "type": "integer",
      "description": "憑證 ID"
    },
    "domain_name": {
      "type": "string",
      "description": "域名名稱"
    }
  },
  "oneOf": [
    {"required": ["certificate_id"]},
    {"required": ["domain_name"]}
  ]
}
```

**Output**:
```json
{
  "success": true,
  "message": "憑證續約已啟動",
  "job_id": "renewal-job-123"
}
```

---

### 2.10 check_dns

檢查域名的 DNS 配置。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "domain": {
      "type": "string",
      "description": "要檢查的域名"
    }
  },
  "required": ["domain"]
}
```

**Output**:
```json
{
  "domain": "example.com",
  "configured": true,
  "records": [
    {
      "type": "A",
      "value": "203.0.113.1",
      "ttl": 300
    }
  ],
  "expected_ip": "203.0.113.1",
  "actual_ips": ["203.0.113.1"],
  "matches": true,
  "message": "DNS 配置正確"
}
```

---

### 2.11 get_diagnostics

取得域名或系統的診斷資訊。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {
    "domain_id": {
      "type": "integer",
      "description": "域名 ID (選填,不提供則返回系統級診斷)"
    },
    "log_type": {
      "type": "string",
      "enum": ["info", "warning", "error"],
      "description": "篩選日誌類型"
    },
    "resolved": {
      "type": "boolean",
      "description": "篩選已解決或未解決的問題"
    },
    "limit": {
      "type": "integer",
      "default": 20,
      "minimum": 1,
      "maximum": 100
    }
  }
}
```

**Output**:
```json
{
  "logs": [
    {
      "id": 1,
      "domain_id": 1,
      "log_type": "warning",
      "category": "certificate",
      "message": "憑證將在 15 天內到期",
      "details": {
        "certificate_id": 1,
        "expires_at": "2025-11-22T10:00:00Z"
      },
      "resolved": false,
      "created_at": "2025-11-07T10:00:00Z"
    }
  ],
  "total": 1
}
```

---

### 2.12 get_system_health

取得系統健康狀態。

**Input Schema**:
```json
{
  "type": "object",
  "properties": {}
}
```

**Output**:
```json
{
  "status": "healthy",
  "timestamp": "2025-11-07T10:00:00Z",
  "components": {
    "database": {
      "status": "healthy",
      "message": "連接正常",
      "latency_ms": 2
    },
    "kubernetes": {
      "status": "healthy",
      "message": "API 伺服器可達",
      "latency_ms": 15
    },
    "cert_manager": {
      "status": "healthy",
      "message": "cert-manager 運行中",
      "latency_ms": 10
    }
  }
}
```

---

## 3. Resources (資源列表)

MCP Resources 提供靜態或動態資料,可被 AI 助理讀取和參考。

### 3.1 domain://list

所有域名列表資源。

**URI**: `domain://list`

**MIME Type**: `application/json`

**Content**:
```json
{
  "domains": [
    {
      "id": 1,
      "domain_name": "example.com",
      "status": "active",
      "target_service": "nginx-service",
      "certificate_status": "valid",
      "certificate_expires": "2026-02-05T10:00:00Z"
    }
  ]
}
```

---

### 3.2 domain://{name}

單一域名資源。

**URI Pattern**: `domain://{domain_name}`

**範例 URI**: `domain://example.com`

**MIME Type**: `application/json`

**Content**:
```json
{
  "id": 1,
  "domain_name": "example.com",
  "target_service": "nginx-service",
  "target_namespace": "default",
  "target_port": 80,
  "ssl_mode": "auto",
  "status": "active",
  "certificate": {
    "id": 1,
    "source": "letsencrypt",
    "valid_until": "2026-02-05T10:00:00Z",
    "status": "valid"
  },
  "diagnostics": [
    {
      "log_type": "info",
      "message": "域名運行正常"
    }
  ]
}
```

---

### 3.3 service://list

Kubernetes Services 列表資源。

**URI**: `service://list`

**MIME Type**: `application/json`

**Content**:
```json
{
  "services": [
    {
      "name": "nginx-service",
      "namespace": "default",
      "type": "ClusterIP",
      "ports": [80, 443],
      "labels": {
        "app": "nginx"
      }
    }
  ]
}
```

---

### 3.4 certificate://list

SSL 憑證列表資源。

**URI**: `certificate://list`

**MIME Type**: `application/json`

**Content**:
```json
{
  "certificates": [
    {
      "id": 1,
      "domain_name": "example.com",
      "source": "letsencrypt",
      "valid_until": "2026-02-05T10:00:00Z",
      "status": "valid",
      "days_until_expiry": 90
    }
  ]
}
```

---

### 3.5 diagnostics://logs

系統診斷日誌資源。

**URI**: `diagnostics://logs`

**Query Parameters**:
- `log_type`: 篩選日誌類型 (info, warning, error)
- `resolved`: 篩選已解決或未解決 (true, false)

**範例 URI**: `diagnostics://logs?log_type=error&resolved=false`

**MIME Type**: `application/json`

**Content**:
```json
{
  "logs": [
    {
      "id": 5,
      "domain_id": 3,
      "log_type": "error",
      "category": "certificate",
      "message": "憑證續約失敗: 速率限制",
      "resolved": false,
      "created_at": "2025-11-07T09:30:00Z"
    }
  ]
}
```

---

## 4. Prompts (提示詞範本)

本系統暫不提供 MCP Prompts,未來可擴展支援。

---

## 5. 錯誤處理

### 5.1 錯誤格式 (JSON-RPC 2.0)

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32600,
    "message": "Invalid Request",
    "data": {
      "details": "Missing required field: domain_name"
    }
  }
}
```

### 5.2 錯誤碼

| Code | 名稱 | 說明 |
|------|------|------|
| `-32700` | Parse error | JSON 解析錯誤 |
| `-32600` | Invalid Request | 請求格式不正確 |
| `-32601` | Method not found | 方法不存在 |
| `-32602` | Invalid params | 參數不正確 |
| `-32603` | Internal error | 內部錯誤 |
| `-32001` | Unauthorized | 認證失敗 |
| `-32002` | Not found | 資源不存在 |
| `-32003` | Validation error | 資料驗證失敗 |
| `-32004` | Conflict | 資源衝突 (如域名已存在) |

---

## 6. 認證機制

### 6.1 API Key 認證

**HTTP Header**:
```
X-API-Key: your-api-key-here
```

**錯誤回應** (缺少或無效 API Key):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32001,
    "message": "Unauthorized",
    "data": {
      "details": "Invalid or missing API key"
    }
  }
}
```

---

## 7. Claude Desktop 整合範例

### 7.1 設定檔案

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

**Linux**: `~/.config/Claude/claude_desktop_config.json`

**設定內容**:
```json
{
  "mcpServers": {
    "k8s-domain-manager": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "X-API-Key": "your-api-key-here"
      },
      "metadata": {
        "name": "Kubernetes Domain Manager",
        "description": "管理 Kubernetes 域名和 SSL 憑證"
      }
    }
  }
}
```

### 7.2 使用範例

**使用者**: 幫我列出所有域名及其憑證狀態

**Claude** (執行):
```json
// 呼叫 list_domains tool
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "list_domains",
    "arguments": {}
  }
}
```

**回應**:
```
您目前有 3 個域名:

1. **example.com**
   - 狀態: 運行中 (active)
   - 目標服務: nginx-service
   - 憑證: 有效 (到期日: 2026-02-05)

2. **api.example.com**
   - 狀態: 運行中 (active)
   - 目標服務: api-service
   - 憑證: 即將到期 (到期日: 2025-12-07,剩 30 天)

3. **admin.example.com**
   - 狀態: 運行中 (active)
   - 目標服務: admin-service
   - 憑證: 有效 (到期日: 2026-01-15)
```

---

**使用者**: api.example.com 的憑證快到期了,幫我續約

**Claude** (執行):
```json
// 呼叫 renew_certificate tool
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "renew_certificate",
    "arguments": {
      "domain_name": "api.example.com"
    }
  }
}
```

**回應**:
```
已為 api.example.com 啟動憑證續約程序。

續約任務 ID: renewal-job-123

續約通常需要 5-10 分鐘完成,您可以稍後透過 `get_certificate_status` 工具查詢憑證狀態。
```

---

**使用者**: 新增一個子域名 blog.example.com,指向 wordpress-service 的 80 埠

**Claude** (執行):
```json
// 先檢查服務是否存在
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "list_services",
    "arguments": {
      "namespace": "default"
    }
  }
}

// 確認服務存在後,建立域名
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "create_domain",
    "arguments": {
      "domain_name": "blog.example.com",
      "target_service": "wordpress-service",
      "target_namespace": "default",
      "target_port": 80,
      "ssl_mode": "auto"
    }
  }
}
```

**回應**:
```
已成功建立域名配置:

- 域名: blog.example.com
- 目標服務: wordpress-service (default namespace, port 80)
- SSL 模式: 自動 (Let's Encrypt)
- 狀態: 待配置 (pending)

系統正在自動配置 Ingress 和申請 SSL 憑證,預計 5-10 分鐘內完成。

請確保您已在 DNS 註冊商將 blog.example.com 的 A 記錄指向叢集的 LoadBalancer IP。
```

---

## 8. 實作建議

### 8.1 Go 實作架構

```go
package mcp

import (
    "encoding/json"
    "net/http"
)

// MCP Server Handler
type MCPServer struct {
    domainService      *service.DomainService
    certificateService *service.CertificateService
    k8sClient          *kubernetes.Client
}

// JSON-RPC 2.0 Request
type JSONRPCRequest struct {
    JSONRPC string          `json:"jsonrpc"`
    ID      interface{}     `json:"id"`
    Method  string          `json:"method"`
    Params  json.RawMessage `json:"params"`
}

// JSON-RPC 2.0 Response
type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *RPCError   `json:"error,omitempty"`
}

// RPC Error
type RPCError struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// HTTP Handler
func (s *MCPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // 驗證 API Key
    apiKey := r.Header.Get("X-API-Key")
    if !s.validateAPIKey(apiKey) {
        s.sendError(w, nil, -32001, "Unauthorized", nil)
        return
    }

    // 解析請求
    var req JSONRPCRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        s.sendError(w, nil, -32700, "Parse error", nil)
        return
    }

    // 路由到對應的工具
    var result interface{}
    var rpcErr *RPCError

    switch req.Method {
    case "tools/call":
        result, rpcErr = s.handleToolCall(req.Params)
    case "resources/read":
        result, rpcErr = s.handleResourceRead(req.Params)
    default:
        rpcErr = &RPCError{Code: -32601, Message: "Method not found"}
    }

    // 發送回應
    if rpcErr != nil {
        s.sendError(w, req.ID, rpcErr.Code, rpcErr.Message, rpcErr.Data)
    } else {
        s.sendResult(w, req.ID, result)
    }
}

// Tool Call Handler
func (s *MCPServer) handleToolCall(params json.RawMessage) (interface{}, *RPCError) {
    var toolCall struct {
        Name      string                 `json:"name"`
        Arguments map[string]interface{} `json:"arguments"`
    }

    if err := json.Unmarshal(params, &toolCall); err != nil {
        return nil, &RPCError{Code: -32602, Message: "Invalid params"}
    }

    switch toolCall.Name {
    case "list_domains":
        return s.listDomains(toolCall.Arguments)
    case "create_domain":
        return s.createDomain(toolCall.Arguments)
    // ... 其他工具
    default:
        return nil, &RPCError{Code: -32601, Message: "Tool not found"}
    }
}
```

---

## 9. 測試策略

### 9.1 單元測試

```go
func TestMCPListDomains(t *testing.T) {
    server := NewMCPServer(mockDomainService, mockCertService, mockK8sClient)

    req := `{
        "jsonrpc": "2.0",
        "id": 1,
        "method": "tools/call",
        "params": {
            "name": "list_domains",
            "arguments": {"limit": 10}
        }
    }`

    w := httptest.NewRecorder()
    r := httptest.NewRequest("POST", "/mcp", strings.NewReader(req))
    r.Header.Set("X-API-Key", "test-api-key")

    server.ServeHTTP(w, r)

    assert.Equal(t, 200, w.Code)

    var resp JSONRPCResponse
    json.NewDecoder(w.Body).Decode(&resp)

    assert.Nil(t, resp.Error)
    assert.NotNil(t, resp.Result)
}
```

### 9.2 整合測試

使用 MCP Inspector 或自訂測試客戶端測試完整流程:

```bash
# 使用 MCP Inspector 測試
npm install -g @modelcontextprotocol/inspector
mcp-inspector http://localhost:8080/mcp --header "X-API-Key: test-key"
```

---

## 10. 規格摘要

### 工具統計

- **總工具數**: 12 個
- **域名管理**: 5 個 (list, get, create, update, delete)
- **憑證管理**: 3 個 (get_status, list_expiring, renew)
- **服務發現**: 1 個 (list_services)
- **診斷**: 3 個 (get_diagnostics, check_dns, get_system_health)

### 資源統計

- **總資源數**: 5 個
- **域名資源**: 2 個 (list, single)
- **服務資源**: 1 個 (list)
- **憑證資源**: 1 個 (list)
- **診斷資源**: 1 個 (logs)

### 設計決策

1. **選擇 HTTP/SSE**: 比 stdio 更易於除錯和遠端連接
2. **JSON-RPC 2.0**: 標準協議,易於實作和測試
3. **API Key 認證**: 與 REST API 統一認證機制
4. **靈活查詢**: 支援按 ID 或名稱查詢,方便 AI 助理使用
5. **豐富錯誤資訊**: 提供詳細錯誤訊息,幫助 AI 理解問題

---

**文件版本**: 1.0
**最後更新**: 2025-11-07
