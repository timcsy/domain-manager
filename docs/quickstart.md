# 快速入門指南

## 前置需求

- Kubernetes 1.20+ 叢集
- Helm 3.0+
- kubectl 已配置連線到叢集
- Nginx Ingress Controller（推薦）

## 步驟 1：安裝

```bash
# 從本地 chart 安裝
helm install domain-manager ./helm/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  --set admin.password="your-secure-password" \
  --set ingress.certManager.issuerEmail="your-email@example.com"
```

## 步驟 2：存取 Web UI

```bash
# Port forward
kubectl port-forward -n domain-manager svc/domain-manager 8080:8080

# 開啟瀏覽器
open http://localhost:8080
```

預設登入：
- 帳號：`admin`
- 密碼：安裝時設定的密碼（預設 `admin`，請務必修改）

## 步驟 3：新增域名

1. 登入後進入「域名管理」頁面
2. 點擊「新增域名」
3. 填入：
   - **域名**：`app.example.com`
   - **目標服務**：從下拉選單選擇 K8s 服務
   - **目標埠號**：`80`
   - **SSL 模式**：`auto`（自動透過 Let's Encrypt）
4. 點擊建立

系統會自動：
- 建立對應的 Kubernetes Ingress 資源
- 透過 Let's Encrypt 申請 SSL 憑證
- 設定 TLS Secret

## 步驟 4：設定 DNS

將你的域名 DNS A 記錄指向 Ingress Controller 的 LoadBalancer IP：

```bash
# 取得 Ingress Controller 的外部 IP
kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
```

## 步驟 5：建立 API Key（選擇性）

如需程式化存取：

1. 進入「API 金鑰」頁面
2. 點擊「建立金鑰」
3. 設定名稱和權限
4. 複製並保存顯示的金鑰

使用方式：
```bash
curl -H "X-API-Key: dm_your_key_here" http://localhost:8080/api/v1/domains
```

## 步驟 6：MCP 整合（選擇性）

如需讓 AI 工具操作域名管理器，請參閱 [MCP 使用範例](mcp-examples.md)。

## 本地開發

不需要 Kubernetes 叢集也能開發：

```bash
# 設定 mock 模式
cd backend
export K8S_MOCK=true
go run src/main.go

# 另一個終端，建置前端 CSS
cd frontend
npm install
npm run watch:css
```

存取 http://localhost:8080，Mock 模式會模擬 K8s 服務和 Ingress 操作。

## 生產環境建議

- 修改預設管理員密碼
- 設定正確的 Let's Encrypt 郵箱
- 啟用持久化儲存（預設已啟用）
- 設定適當的資源限制
- 定期備份（系統預設每日凌晨 2 點自動備份）
