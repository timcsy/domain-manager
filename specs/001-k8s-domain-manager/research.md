# 技術研究報告: Kubernetes 域名管理器

**專案**: domain-manager
**功能分支**: 001-k8s-domain-manager
**研究日期**: 2025-11-07
**研究目標**: 為 Kubernetes 域名管理器選擇最佳技術棧,重點為「簡單優先、資源高效、成熟穩定」

---

## 1. 後端語言選擇

### 決策: **Go (Golang)**

#### 理由

1. **K8s 生態系統原生支援**
   - Go 是 Kubernetes 的官方開發語言,約 75% 的 CNCF 專案使用 Go
   - client-go 是最成熟、功能最完整的 K8s 客戶端庫
   - 與 K8s API 無縫整合,文件完善

2. **資源使用極佳**
   - 記憶體佔用比 Python 低 50-70%
   - 啟動時間更快(0.8s vs Python 1.1s)
   - CPU 使用效率高,適合資源受限環境(≤512MB 記憶體限制)

3. **容器映像大小**
   - 使用 distroless 基礎映像可達到 9-10MB
   - 靜態編譯後的單一執行檔,無外部依賴
   - 部署簡單、啟動快速

4. **併發模型優勢**
   - Goroutines 輕量級協程,非常適合處理多個域名的並發操作
   - 內建併發模型,處理 Let's Encrypt API 呼叫、健康檢查等異步任務時無需額外框架

#### 技術細節

**K8s 客戶端**: `client-go` (官方庫)
- 版本: v0.29.0+
- 功能: 完整支援所有 K8s API,包含 Ingress、Service、Secret 管理
- 範例:
  ```go
  import (
      "k8s.io/client-go/kubernetes"
      "k8s.io/client-go/rest"
  )

  config, _ := rest.InClusterConfig()
  clientset, _ := kubernetes.NewForConfig(config)
  ```

**Let's Encrypt 客戶端**: `lego` (go-acme/lego)
- 版本: v4.15.0+
- 優勢:
  - 純 Go 實作,靜態編譯無依賴
  - 支援 150+ DNS 供應商
  - 同時支援 HTTP-01 和 DNS-01 挑戰
  - 文件完善,社群活躍
- 範例:
  ```go
  import "github.com/go-acme/lego/v4/certificate"
  ```

**Web 框架**: `Chi` (go-chi/chi)
- 版本: v5.0.0+
- 選擇理由:
  - 輕量級(僅路由功能,無多餘功能)
  - 與標準庫 `net/http` 100% 相容
  - 中介軟體機制簡單清晰
  - 效能優異,記憶體佔用低
  - 相較於 Gin 和 Echo,Chi 最輕量且易於理解
- 替代方案考慮:
  - Gin: 功能豐富但略顯複雜
  - Echo: 效能最佳但框架較重

**SQLite 庫**: `modernc.org/sqlite`
- 版本: v1.28.0+
- 選擇理由:
  - **CGo-free** (無 C 依賴),靜態編譯友善
  - 跨平台編譯簡單
  - 雖然效能比 mattn/go-sqlite3 慢 10-200%,但對本專案(≤100 域名)影響可忽略
  - 避免 CGo 帶來的編譯複雜度和安全問題
- 效能考量:
  - 簡單查詢慢 10-20%
  - 批次插入慢約 2 倍
  - 對於管理介面的低頻操作,差異不明顯

**預估資源使用**:
- 記憶體: 30-80MB (空載時 ~30MB,管理 100 個域名時 ~80MB)
- CPU: 0.1-0.3 核心(日常運行),峰值 0.5 核心(證書申請時)
- 容器映像: 10-15MB (使用 distroless 或 scratch 基礎映像)

#### 替代方案考慮: Python

**為何未選擇 Python**:
1. **記憶體佔用高**: FastAPI 應用基礎記憶體佔用 100-150MB,不符合 ≤512MB 限制
2. **容器映像大**: 最小化 Python 映像約 50-80MB,明顯大於 Go
3. **效能**: 同步操作效率低,異步框架(FastAPI)增加複雜度
4. **依賴管理**: 需要 pip、虛擬環境,部署較複雜

**Python 優勢**:
- 開發速度快
- certbot/acme 生態成熟
- kubernetes-client 功能完整

**結論**: 對於資源受限的 K8s 環境,Go 更適合

---

## 2. 前端技術棧

### 決策: **HTMX + TailwindCSS + Alpine.js (輕量混合方案)**

#### 理由

1. **極小打包大小**
   - HTMX: ~14KB (gzipped)
   - Alpine.js: ~15KB (gzipped)
   - TailwindCSS: 產生後僅包含使用的樣式(~10-30KB)
   - 總計: < 60KB (相比 React/Vue 200KB+)

2. **開發效率**
   - HTMX: 伺服器端渲染,無需複雜的前後端分離架構
   - Alpine.js: 處理客戶端互動(模態框、下拉選單、即時驗證)
   - TailwindCSS: 快速構建一致的 UI,無需手寫 CSS

3. **複雜度低**
   - 無需 Node.js 構建工具鏈
   - 無需學習複雜的狀態管理(Redux、Vuex)
   - 伺服器端直接渲染 HTML,易於除錯

4. **適合專案需求**
   - 管理介面為表單驅動(新增/編輯域名、上傳證書)
   - 不需要複雜的即時互動或大量客戶端狀態
   - HTMX 的 AJAX 更新非常適合表單提交和資料刷新

#### 技術細節

**架構**:
```
Go (Chi) → 伺服器端渲染 HTML → HTMX (AJAX 互動) + Alpine.js (客戶端狀態)
```

**HTMX 使用場景**:
- 表單提交(新增域名、上傳證書)
- 分頁和篩選(域名列表)
- 動態載入(服務列表、診斷資訊)

**Alpine.js 使用場景**:
- 模態框開關
- 下拉選單展開/收合
- 表單即時驗證
- Toast 通知

**TailwindCSS**:
- 使用 CDN 版本(開發階段)
- 產品版本使用 Tailwind CLI 產生最小化 CSS

**範例程式碼**:
```html
<!-- HTMX: 點擊按鈕發送 POST 請求,更新目標區域 -->
<button hx-post="/api/domains"
        hx-target="#domain-list"
        hx-swap="outerHTML">
  新增域名
</button>

<!-- Alpine.js: 控制模態框 -->
<div x-data="{ open: false }">
  <button @click="open = true">開啟設定</button>
  <div x-show="open" @click.away="open = false">
    模態框內容
  </div>
</div>
```

#### 替代方案考慮

**選項 A: React/Vue/Svelte + TypeScript**
- 優勢: 功能強大、適合複雜 SPA
- 劣勢:
  - 打包大小 200KB+
  - 需要 Node.js 構建工具
  - 學習曲線陡峭
  - 對本專案過度工程

**選項 C: 純 HTML + 極簡 JavaScript**
- 優勢: 最輕量
- 劣勢: 無框架支援,開發效率低,程式碼可維護性差

**結論**: HTMX + Alpine.js 是「簡單」與「功能」的最佳平衡

---

## 3. Kubernetes Ingress 控制器整合

### 決策: **通用 Ingress 資源操作 + Nginx Ingress Controller 優先支援**

#### 整合策略

**核心原則**: 直接操作標準 Kubernetes Ingress 資源,保持控制器無關性

**實作方式**:
1. **使用標準 Ingress API**
   - 透過 client-go 建立/更新 `networking.k8s.io/v1/Ingress` 資源
   - 所有主流 Ingress 控制器都支援標準 Ingress 規格

2. **預設支援 Nginx Ingress Controller**
   - Nginx Ingress 是最流行的選擇,社群最大
   - 在 Helm Chart values.yaml 提供 Nginx 特定註解範本
   - 範例:
     ```yaml
     metadata:
       annotations:
         kubernetes.io/ingress.class: nginx
         cert-manager.io/cluster-issuer: letsencrypt-prod
         nginx.ingress.kubernetes.io/ssl-redirect: "true"
     ```

3. **相容其他控制器**
   - Traefik: 支援標準 Ingress + 自訂 IngressRoute CRD(進階功能)
   - 使用者可在 Helm values.yaml 覆蓋註解

**Ingress 資源範例**:
```go
ingress := &networkingv1.Ingress{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "example-com",
        Namespace: "default",
        Annotations: map[string]string{
            "kubernetes.io/ingress.class":              "nginx",
            "cert-manager.io/cluster-issuer":           "letsencrypt-prod",
            "nginx.ingress.kubernetes.io/ssl-redirect": "true",
        },
    },
    Spec: networkingv1.IngressSpec{
        TLS: []networkingv1.IngressTLS{{
            Hosts:      []string{"example.com"},
            SecretName: "example-com-tls",
        }},
        Rules: []networkingv1.IngressRule{{
            Host: "example.com",
            IngressRuleValue: networkingv1.IngressRuleValue{
                HTTP: &networkingv1.HTTPIngressRuleValue{
                    Paths: []networkingv1.HTTPIngressPath{{
                        Path:     "/",
                        PathType: &pathTypePrefix,
                        Backend: networkingv1.IngressBackend{
                            Service: &networkingv1.IngressServiceBackend{
                                Name: "my-service",
                                Port: networkingv1.ServiceBackendPort{
                                    Number: 80,
                                },
                            },
                        },
                    }},
                },
            },
        }},
    },
}
```

#### Nginx vs Traefik 考量

**Nginx Ingress Controller**:
- 優勢: 成熟穩定、社群最大、文件完善、適合初學者
- 劣勢: 需手動配置,不支援動態服務發現

**Traefik**:
- 優勢: 自動服務發現、原生支援 Let's Encrypt、Web UI
- 劣勢: 配置較複雜、社群較小

**結論**: 優先支援 Nginx,但保持標準 Ingress API 以相容其他控制器

---

## 4. Let's Encrypt 整合

### 決策: **cert-manager 整合(推薦) + 自行實作 lego 作為備援**

#### 推薦方案: cert-manager

**選擇理由**:
1. **成熟穩定**: K8s 事實標準,經過大規模生產驗證
2. **自動化程度高**: 自動偵測 Ingress、申請/續約證書
3. **效能優異**: 比早期版本減少 100 倍 ACME API 呼叫
4. **支援多種挑戰**: HTTP-01、DNS-01、TLS-ALPN-01
5. **簡化架構**: 無需自行處理證書生命週期

**整合方式**:
1. **前置條件檢查**
   - 在 Helm Chart 安裝時檢測 cert-manager 是否已安裝
   - 若未安裝,提供安裝指引或自動安裝(可選)

2. **建立 ClusterIssuer**
   ```yaml
   apiVersion: cert-manager.io/v1
   kind: ClusterIssuer
   metadata:
     name: letsencrypt-prod
   spec:
     acme:
       server: https://acme-v02.api.letsencrypt.org/directory
       email: user@example.com  # 從 Helm values.yaml 注入
       privateKeySecretRef:
         name: letsencrypt-prod
       solvers:
       - http01:
           ingress:
             class: nginx
   ```

3. **Ingress 註解自動化**
   - 系統自動在 Ingress 加入 `cert-manager.io/cluster-issuer: letsencrypt-prod`
   - cert-manager 自動處理證書申請和 Secret 建立

**憑證續約**:
- cert-manager 自動在到期前 30 天續約
- 系統僅需監控證書狀態並在管理介面顯示

#### 備援方案: 自行實作 lego

**使用時機**:
- cert-manager 未安裝且使用者不想安裝
- 需要完全控制證書流程
- 特殊 DNS 供應商需求

**實作重點**:
1. **HTTP-01 挑戰** (優先)
   - 在 Ingress 建立 `/.well-known/acme-challenge/` 路徑
   - lego 處理 ACME 伺服器請求
   - 取得證書後建立 K8s Secret

2. **DNS-01 挑戰** (進階)
   - 支援萬用字元證書
   - 需要 DNS 供應商 API 整合
   - 複雜度高,非核心功能

3. **證書續約**
   - 背景 Goroutine 每日檢查證書到期時間
   - 到期前 30 天自動續約

**程式碼範例**:
```go
import (
    "github.com/go-acme/lego/v4/certificate"
    "github.com/go-acme/lego/v4/lego"
)

client, _ := lego.NewClient(config)
request := certificate.ObtainRequest{
    Domains: []string{"example.com"},
    Bundle:  true,
}
certificates, _ := client.Certificate.Obtain(request)
```

#### HTTP-01 vs DNS-01 挑戰比較

| 特性 | HTTP-01 | DNS-01 |
|------|---------|--------|
| **複雜度** | 簡單 | 複雜(需 DNS API) |
| **萬用字元證書** | ❌ 不支援 | ✅ 支援 |
| **埠口需求** | 需開放 80 埠 | 無埠口需求 |
| **推薦使用** | cert-manager | cert-manager |
| **適用場景** | 標準域名 | *.example.com |

**結論**: HTTP-01 + cert-manager 為預設方案,簡單且符合 99% 使用場景

---

## 5. MCP (Model Context Protocol) 實作

### 實作指引

#### 選擇 SDK: **Go 自行實作 MCP 協議**

**理由**:
1. **官方 Go SDK 尚不成熟**: 雖然 MCP 有官方 Go SDK,但功能較 Python/TypeScript 版本滯後
2. **協議簡單**: MCP 基於 JSON-RPC 2.0,直接實作不困難
3. **統一語言**: 避免引入 Python/TypeScript 依賴,保持專案純 Go

#### MCP 伺服器架構

**協議**: JSON-RPC 2.0 over stdio/HTTP/WebSocket

**實作方式**: HTTP/SSE (Server-Sent Events)
- 較 stdio 更易於除錯和測試
- 支援遠端連接(Claude Desktop 可透過 HTTP 連接)

**提供的 MCP 工具** (Tools):
1. `list_domains`: 列出所有域名及狀態
2. `get_domain`: 取得單一域名詳細資訊
3. `create_domain`: 新增域名配置
4. `update_domain`: 更新域名配置
5. `delete_domain`: 刪除域名配置
6. `list_services`: 列出可用的 K8s 服務
7. `get_certificate_status`: 查詢證書狀態

**提供的 MCP 資源** (Resources):
- `domain://list`: 域名列表資源
- `domain://{name}`: 單一域名資源
- `service://list`: 服務列表資源

**範例程式碼**:
```go
// MCP Tool Definition
type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCP Server Handler
func handleToolCall(toolName string, args map[string]interface{}) (interface{}, error) {
    switch toolName {
    case "list_domains":
        return domainService.ListDomains()
    case "create_domain":
        domain := args["domain"].(string)
        service := args["service"].(string)
        return domainService.CreateDomain(domain, service)
    // ... 其他工具
    }
}
```

#### 與 REST API 整合

**架構**:
```
                ┌─────────────────┐
                │   Go Backend    │
                │                 │
                │  ┌───────────┐  │
                │  │  Domain   │  │
                │  │  Service  │  │
                │  │  (Core)   │  │
                │  └─────┬─────┘  │
                │        │        │
                │   ┌────┴────┐   │
                │   │         │   │
          ┌─────┼───┤         ├───┼─────┐
          │     │   │         │   │     │
     ┌────▼──┐  │ ┌─▼───┐ ┌───▼─┐ │ ┌───▼────┐
     │ REST  │  │ │ Web │ │ MCP │ │ │ K8s    │
     │ API   │  │ │ UI  │ │ API │ │ │ Client │
     └───────┘  │ └─────┘ └─────┘ │ └────────┘
                └─────────────────┘
```

**共用業務邏輯**:
- REST API、Web UI、MCP Server 都呼叫同一個 `DomainService`
- 認證機制: API Key(REST/MCP) 或 Session Cookie(Web UI)

**MCP 身份驗證**:
```go
// MCP 請求驗證
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        apiKey := r.Header.Get("X-API-Key")
        if !isValidAPIKey(apiKey) {
            http.Error(w, "Unauthorized", 401)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

#### Claude Desktop 整合範例

**設定檔** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "k8s-domain-manager": {
      "url": "http://localhost:8080/mcp",
      "headers": {
        "X-API-Key": "your-api-key"
      }
    }
  }
}
```

**使用範例**:
```
使用者: 幫我列出所有域名及其證書狀態
Claude: [呼叫 list_domains 工具]
       您目前有 3 個域名:
       1. example.com - 證書有效(到期: 2025-03-01)
       2. api.example.com - 證書即將到期(到期: 2025-01-15)
       3. admin.example.com - 證書有效(到期: 2025-02-20)
```

#### 參考資源

- MCP 規範: https://modelcontextprotocol.io/
- MCP Go SDK: https://github.com/modelcontextprotocol/go-sdk (參考用)
- 範例實作: https://github.com/modelcontextprotocol/servers

---

## 6. 測試策略

### 推薦工具鏈

#### 單元測試

**框架**: Go 標準庫 `testing` + `testify/assert`

**選擇理由**:
- 標準庫足夠簡單且功能完整
- testify/assert 提供更友善的斷言語法
- 無過度依賴,保持輕量

**範例**:
```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestDomainValidation(t *testing.T) {
    assert.True(t, isValidDomain("example.com"))
    assert.False(t, isValidDomain("invalid domain"))
}
```

**測試覆蓋率目標**: ≥ 70%

**重點測試項目**:
- 域名驗證邏輯
- SQLite CRUD 操作
- ACME 挑戰處理
- Ingress 資源建立

#### 整合測試

**框架**: `testcontainers-go` + `kind` (Kubernetes in Docker)

**測試內容**:
1. **資料庫整合**
   - 使用 testcontainers 啟動真實 SQLite
   - 測試遷移、查詢、事務

2. **Kubernetes 整合**
   - 使用 kind 建立測試叢集
   - 測試 Ingress/Secret/Service 操作
   - 驗證 RBAC 權限

**範例**:
```go
func TestCreateIngress(t *testing.T) {
    // 啟動 kind 叢集
    cluster := kind.NewCluster()
    defer cluster.Delete()

    // 建立 Ingress
    err := createIngress("example.com", "my-service")
    assert.NoError(t, err)

    // 驗證 Ingress 存在
    ingress, _ := getIngress("example-com")
    assert.Equal(t, "example.com", ingress.Spec.Rules[0].Host)
}
```

#### E2E 測試

**工具**: `playwright-go` 或手動測試腳本

**測試場景**:
1. **安裝流程**
   - `helm install` → 存取管理介面 → 首次登入

2. **域名配置流程**
   - 登入 → 新增域名 → 選擇服務 → 申請證書 → 驗證 HTTPS 可用

3. **證書管理**
   - 上傳自訂證書 → 檢視狀態 → 刪除域名

**執行方式**:
- CI/CD: 使用 GitHub Actions 在真實 K8s 叢集測試
- 本地: 使用 kind 或 minikube

**範例 GitHub Actions**:
```yaml
name: E2E Tests
on: [push]
jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: helm/kind-action@v1.5.0
      - run: |
          helm install domain-manager ./chart
          kubectl wait --for=condition=ready pod -l app=domain-manager
          # 執行 E2E 測試腳本
```

#### Mock 和存根

**工具**: `gomock` 或 `testify/mock`

**使用場景**:
- Mock K8s API 回應
- Mock Let's Encrypt ACME 伺服器
- Mock DNS 解析

**範例**:
```go
type MockK8sClient struct {
    mock.Mock
}

func (m *MockK8sClient) CreateIngress(ingress *v1.Ingress) error {
    args := m.Called(ingress)
    return args.Error(0)
}
```

#### CI/CD 整合

**平台**: GitHub Actions

**流程**:
```
提交程式碼 → 單元測試 → 整合測試 → 建置容器映像 → E2E 測試 → 發布
```

**測試分層**:
- **快速回饋**: 單元測試 (<30 秒)
- **中等回饋**: 整合測試 (2-5 分鐘)
- **完整驗證**: E2E 測試 (10-15 分鐘,僅在 main 分支)

---

## 7. 技術棧總結

| 類別 | 技術選擇 | 版本 | 原因 |
|------|---------|------|------|
| **後端語言** | Go | 1.22+ | 資源高效、K8s 原生、容器映像小 |
| **Web 框架** | Chi | v5.0+ | 輕量、標準庫相容、簡單 |
| **K8s 客戶端** | client-go | v0.29+ | 官方庫、功能完整 |
| **Let's Encrypt** | lego | v4.15+ | 純 Go、無依賴、支援廣泛 |
| **證書管理** | cert-manager | v1.14+ | K8s 標準、自動化、成熟 |
| **SQLite** | modernc.org/sqlite | v1.28+ | CGo-free、靜態編譯 |
| **前端框架** | HTMX + Alpine.js | 1.9+ / 3.13+ | 極輕量、開發簡單 |
| **CSS 框架** | TailwindCSS | 3.4+ | 快速開發、小打包大小 |
| **Ingress 控制器** | Nginx Ingress | 任意版本 | 最流行、文件完善 |
| **MCP 實作** | 自行實作 | - | 保持純 Go、協議簡單 |
| **測試框架** | testing + testify | stdlib + v1.9+ | 標準 + 易用斷言 |
| **容器基礎映像** | distroless | latest | 最小大小、最佳安全性 |

---

## 8. 資源預算驗證

基於以上技術選擇,預估資源使用:

| 資源 | 預估使用 | 限制 | 符合? |
|------|---------|------|-------|
| **記憶體** | 30-80MB | ≤512MB | ✅ (佔用 6-16%) |
| **CPU** | 0.1-0.5 核心 | ≤1 核心 | ✅ (佔用 10-50%) |
| **容器映像** | 10-15MB | 無限制 | ✅ (極小) |
| **啟動時間** | <2 秒 | 無限制 | ✅ (快速) |

**結論**: 所有技術選擇均符合資源限制,並有充足餘裕

---

## 9. 風險與緩解

| 風險 | 影響 | 緩解措施 |
|------|------|---------|
| cert-manager 未安裝 | 無法自動證書管理 | 提供安裝指引 + lego 備援方案 |
| Ingress 控制器不相容 | 路由無法建立 | 支援標準 Ingress API,提供多種註解範本 |
| Let's Encrypt 速率限制 | 證書申請失敗 | 顯示清晰錯誤訊息,建議使用自訂證書 |
| SQLite 資料庫損壞 | 配置遺失 | 提供備份/還原功能,建議定期備份 |
| Go 靜態編譯問題 | CGo 依賴導致跨平台編譯失敗 | 使用 CGo-free 庫 (modernc.org/sqlite) |

---

## 10. 下一步行動

1. **建立專案骨架**
   - 初始化 Go 模組
   - 設定專案結構 (MVC 模式)
   - 建立 Makefile 和 Dockerfile

2. **實作核心功能**
   - K8s 客戶端封裝
   - SQLite 資料庫 schema
   - Domain Service (CRUD)

3. **建立 Helm Chart**
   - Chart.yaml、values.yaml
   - RBAC 配置
   - Deployment/Service 範本

4. **實作前端**
   - HTMX + Alpine.js 模板
   - TailwindCSS 樣式

5. **整合測試**
   - 單元測試
   - kind 整合測試
   - E2E 測試

---

## 附錄: 參考資源

### 官方文件
- Go client-go: https://github.com/kubernetes/client-go
- lego: https://go-acme.github.io/lego/
- cert-manager: https://cert-manager.io/
- HTMX: https://htmx.org/
- Alpine.js: https://alpinejs.dev/
- MCP: https://modelcontextprotocol.io/

### 最佳實踐
- Kubernetes Ingress: https://kubernetes.io/docs/concepts/services-networking/ingress/
- Go 容器映像優化: https://docs.docker.com/language/golang/build-images/
- Let's Encrypt 速率限制: https://letsencrypt.org/docs/rate-limits/

---

**報告結論**: 以上技術棧完全符合「簡單優先、開箱即用、資源高效」的專案目標,所有選擇均為成熟穩定的方案,並經過生產環境驗證。建議直接採用此技術棧進行開發。
