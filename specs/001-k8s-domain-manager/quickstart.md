# 快速入門指南: Kubernetes 域名管理器

**功能分支**: `001-k8s-domain-manager`
**建立日期**: 2025-11-07
**狀態**: Draft
**版本**: 1.0

---

## 目標讀者

本指南適合:
- 擁有 Kubernetes 叢集但不熟悉其複雜配置的使用者
- 想要快速設定域名和 SSL 憑證的開發者
- 使用 Vultr、DigitalOcean、Linode 等雲端平台的使用者

---

## 前置需求

在開始之前,請確認您具備:

1. **Kubernetes 叢集**
   - 版本: 1.20 或更高
   - 可透過 `kubectl` 存取
   - 至少有 1 CPU 核心和 512MB 記憶體可用

2. **Helm 3**
   - 版本: 3.0 或更高
   - 安裝指引: https://helm.sh/docs/intro/install/

3. **Ingress 控制器** (推薦)
   - Nginx Ingress Controller (推薦)
   - Traefik 或其他標準 Ingress 控制器

4. **域名**
   - 已在域名註冊商購買的域名
   - 有權限修改 DNS 設定

5. **LoadBalancer 或 NodePort** (可選)
   - 叢集能夠對外提供服務

---

## 安裝步驟

### 步驟 1: 檢查環境

驗證您的 Kubernetes 叢集和工具:

```bash
# 檢查 kubectl 連接
kubectl cluster-info

# 檢查 Helm 版本
helm version

# 檢查節點狀態
kubectl get nodes
```

預期輸出:
```
NAME                 STATUS   ROLES    AGE   VERSION
k8s-worker-1         Ready    <none>   30d   v1.28.0
```

---

### 步驟 2: (可選) 安裝 Nginx Ingress Controller

如果您的叢集尚未安裝 Ingress 控制器:

```bash
# 使用 Helm 安裝 Nginx Ingress
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install nginx-ingress ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --set controller.service.type=LoadBalancer
```

等待 LoadBalancer 分配外部 IP:

```bash
kubectl get svc -n ingress-nginx nginx-ingress-ingress-nginx-controller
```

預期輸出:
```
NAME                                     TYPE           EXTERNAL-IP      PORT(S)
nginx-ingress-ingress-nginx-controller   LoadBalancer   203.0.113.1      80:31234/TCP,443:31235/TCP
```

**記下 EXTERNAL-IP (例如 203.0.113.1),稍後會用到。**

---

### 步驟 3: (可選) 安裝 cert-manager

cert-manager 可以自動管理 Let's Encrypt 憑證(推薦):

```bash
# 安裝 cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# 驗證安裝
kubectl get pods -n cert-manager
```

預期輸出:
```
NAME                                      READY   STATUS    RESTARTS   AGE
cert-manager-7d9c8c9f8d-xxxxx             1/1     Running   0          1m
cert-manager-cainjector-5b9c8c9f8d-xxxxx  1/1     Running   0          1m
cert-manager-webhook-5b9c8c9f8d-xxxxx     1/1     Running   0          1m
```

---

### 步驟 4: 新增 Helm Repository

```bash
# 新增域名管理器 Helm repository (假設已發布)
helm repo add domain-manager https://your-repo-url.com/charts
helm repo update

# 或使用本地 Chart (開發階段)
# 跳過此步驟,直接使用本地路徑
```

---

### 步驟 5: 準備設定檔案

建立 `values.yaml` 自訂安裝設定:

```yaml
# values.yaml
admin:
  # 初始管理員密碼 (請務必修改!)
  password: "your-secure-password-here"
  email: "admin@example.com"  # 用於 Let's Encrypt 通知

# Let's Encrypt 配置 (內建 ACME 客戶端)
letsencrypt:
  email: "admin@example.com"           # 用於憑證到期通知
  staging: false                       # 測試環境請設為 true
  accountPath: "/data/letsencrypt"     # 帳戶資料儲存路徑

ingress:
  # Ingress Class 名稱
  className: "nginx"
  # 如果使用 cert-manager (可選，與內建 Let's Encrypt 二選一)
  certManager:
    enabled: false                     # 如使用內建 Let's Encrypt 請設為 false
    issuerEmail: "admin@example.com"

persistence:
  enabled: true
  size: 1Gi
  # 儲存類別 (根據您的雲端平台調整)
  # storageClass: "standard"  # GKE
  # storageClass: "do-block-storage"  # DigitalOcean
  # storageClass: "vultr-block-storage"  # Vultr

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

# MCP 伺服器配置
mcp:
  enabled: true

# 自動備份配置
backup:
  enabled: true
  retentionDays: 30
```

---

### 步驟 6: 安裝域名管理器

```bash
# 使用 Helm 安裝 (使用遠端 repository)
helm install domain-manager domain-manager/domain-manager \
  --namespace domain-manager \
  --create-namespace \
  --values values.yaml

# 或使用本地 Chart (開發階段)
helm install domain-manager ./chart \
  --namespace domain-manager \
  --create-namespace \
  --values values.yaml
```

安裝通常需要 2-5 分鐘。

---

### 步驟 7: 驗證安裝

```bash
# 檢查 Pod 狀態
kubectl get pods -n domain-manager

# 預期輸出
NAME                              READY   STATUS    RESTARTS   AGE
domain-manager-7d9c8c9f8d-xxxxx   1/1     Running   0          2m
```

```bash
# 檢查服務
kubectl get svc -n domain-manager

# 預期輸出
NAME             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)
domain-manager   ClusterIP   10.96.100.50    <none>        8080/TCP
```

---

### 步驟 8: 存取管理介面

有兩種方式存取:

#### 方式 A: Port Forward (本地測試)

```bash
kubectl port-forward -n domain-manager svc/domain-manager 8080:8080
```

然後在瀏覽器開啟: http://localhost:8080

#### 方式 B: 透過 Ingress (生產環境)

如果您想透過域名存取管理介面,可以建立 Ingress:

```bash
# 在 DNS 註冊商將 admin.yourdomain.com 指向 LoadBalancer IP (203.0.113.1)

# 然後透過 Web UI 或 API 新增域名配置
# 域名: admin.yourdomain.com
# 目標服務: domain-manager (namespace: domain-manager)
# 目標埠: 8080
```

---

## 首次設定

### 1. 登入管理介面

開啟 http://localhost:8080 (或您的域名)

- **使用者名稱**: `admin`
- **密碼**: (您在 values.yaml 中設定的密碼)

首次登入後,系統會顯示快速入門精靈。

---

### 2. 設定 Let's Encrypt 郵箱 (如使用 cert-manager)

在 **設定 > 系統設定** 中:

- **Let's Encrypt Email**: 輸入您的郵箱 (用於證書到期通知)
- **ACME Server**: 保持預設 (生產環境使用)

測試階段可使用 Let's Encrypt Staging 環境:
```
https://acme-staging-v02.api.letsencrypt.org/directory
```

**點擊 [儲存設定]**

---

### 3. 建立 API 金鑰 (可選)

如果您需要使用 REST API 或 MCP:

1. 前往 **設定 > API 金鑰**
2. 點擊 **[建立新金鑰]**
3. 輸入金鑰名稱 (例如 "CLI Tool")
4. 選擇權限 (read, write, delete)
5. 點擊 **[建立]**

**請妥善保管金鑰,系統僅顯示一次!**

---

## 新增第一個域名

### 準備工作

1. **確認您有一個域名** (例如 example.com)
2. **記下 LoadBalancer 的外部 IP** (例如 203.0.113.1)
3. **確認叢集中有一個運行中的服務** (例如 nginx-service)

---

### 步驟 1: 配置 DNS

在您的域名註冊商 (如 GoDaddy, Namecheap, Cloudflare) 設定 DNS 記錄:

| 類型 | 名稱 | 值 | TTL |
|------|------|-------|-----|
| A | @ | 203.0.113.1 | 300 |
| A | www | 203.0.113.1 | 300 |

**說明**:
- `@` 代表根域名 (example.com)
- `www` 代表 www.example.com
- `203.0.113.1` 是您的 LoadBalancer IP
- TTL 300 秒 (5 分鐘)

---

### 步驟 2: 在管理介面新增域名

1. 點擊左側選單 **[域名]**
2. 點擊 **[新增域名]** 按鈕
3. 填寫表單:

   - **域名名稱**: `example.com`
   - **目標服務**: 從下拉選單選擇 `nginx-service` (系統自動發現)
   - **命名空間**: `default` (通常是預設值)
   - **目標埠**: `80` (根據服務的實際埠)
   - **SSL 模式**: `自動 (Let's Encrypt)` (推薦)

4. 點擊 **[建立]**

---

### 步驟 3: 等待配置完成

系統會自動執行:

1. 驗證 DNS 配置 (約 1-2 分鐘)
2. 建立 Kubernetes Ingress 資源
3. 申請 Let's Encrypt SSL 憑證 (約 5-10 分鐘)

**進度顯示**:
- 域名狀態會從 `待配置` → `運行中`
- 憑證狀態會從 `申請中` → `有效`

---

### 步驟 4: 驗證運作

在瀏覽器開啟: https://example.com

您應該會看到:
- HTTPS 連線 (綠色鎖頭圖示)
- 憑證由 Let's Encrypt 簽發
- 頁面顯示您的 nginx-service 內容

---

## 新增子域名

### 範例: 新增 api.example.com

1. **DNS 設定** (在域名註冊商):

   | 類型 | 名稱 | 值 | TTL |
   |------|------|-------|-----|
   | A | api | 203.0.113.1 | 300 |

2. **在管理介面新增**:
   - 域名名稱: `api.example.com`
   - 目標服務: `api-service`
   - 命名空間: `default`
   - 目標埠: `8080`
   - SSL 模式: `自動 (Let's Encrypt)`

3. **等待配置** (約 5-10 分鐘)

4. **驗證**: 開啟 https://api.example.com

---

## 上傳自訂憑證 (進階)

如果您不想使用 Let's Encrypt,可以上傳自己的憑證:

### 步驟 1: 準備憑證檔案

確保您有:
- `certificate.pem` (憑證檔案)
- `private-key.pem` (私鑰檔案)

### 步驟 2: 上傳憑證

1. 前往 **憑證 > 上傳憑證**
2. 域名名稱: `example.com`
3. 上傳 `certificate.pem`
4. 上傳 `private-key.pem`
5. 點擊 **[上傳]**

### 步驟 3: 關聯到域名

1. 編輯域名 `example.com`
2. SSL 模式: 改為 `手動 (自訂憑證)`
3. 選擇憑證: 從下拉選單選擇剛上傳的憑證
4. 點擊 **[儲存]**

---

## 常見問題 (FAQ)

### Q1: 域名狀態一直是 "待配置",怎麼辦?

**可能原因**:
1. DNS 尚未生效 (通常需要 5-30 分鐘)
2. LoadBalancer IP 設定錯誤
3. 目標服務不存在或無法存取

**解決方法**:
```bash
# 1. 檢查 DNS 是否正確
nslookup example.com

# 2. 檢查服務是否存在
kubectl get svc -n default nginx-service

# 3. 查看診斷日誌
# 在管理介面: 域名 > example.com > 診斷
```

---

### Q2: SSL 憑證申請失敗

**可能原因**:
1. DNS 未正確指向叢集
2. Let's Encrypt 速率限制 (每週 5 次失敗嘗試)
3. 防火牆阻擋 80/443 埠

**解決方法**:
```bash
# 1. 驗證 DNS
dig example.com

# 2. 查看憑證錯誤訊息
# 管理介面: 憑證 > example.com > 查看詳情

# 3. 切換到 Let's Encrypt Staging 環境測試
# 設定 > 系統設定 > ACME Server
# 改為: https://acme-staging-v02.api.letsencrypt.org/directory
```

---

### Q3: 如何備份資料?

**方法 1: 透過管理介面**
1. 前往 **設定 > 備份**
2. 點擊 **[建立備份]**
3. 下載生成的備份檔案

**方法 2: 透過 kubectl**
```bash
# 複製資料庫檔案
kubectl cp domain-manager/domain-manager-pod-name:/data/database.db ./backup.db
```

**方法 3: 透過 REST API**
```bash
curl -X POST http://localhost:8080/api/v1/backup \
  -H "X-API-Key: your-api-key" \
  -o backup.db.gz
```

---

### Q4: 如何新增第二個管理員?

目前版本僅支援單一管理員帳戶。如需多使用者支援,請等待後續版本。

**替代方案**: 建立多個 API 金鑰,分配給不同團隊成員使用。

---

### Q5: 如何刪除域名?

1. 前往 **域名** 列表
2. 點擊域名旁的 **[刪除]** 按鈕
3. 確認刪除

**注意**: 預設為軟刪除,域名會被標記為 "已刪除" 但不會從資料庫移除。如需永久刪除,請使用 API 並加上 `hard=true` 參數。

---

### Q6: 憑證即將到期,會自動續約嗎?

是的! 系統會:
1. 在憑證到期前 30 天開始監控
2. 在到期前嘗試自動續約
3. 如果續約失敗,會在管理介面顯示警告

您也可以手動觸發續約:
1. 前往 **憑證** 列表
2. 找到憑證,點擊 **[續約]** 按鈕

---

### Q7: 如何查看系統日誌?

**方法 1: 透過 kubectl**
```bash
kubectl logs -n domain-manager deployment/domain-manager -f
```

**方法 2: 透過管理介面**
1. 前往 **診斷 > 系統日誌**
2. 篩選日誌類型 (資訊/警告/錯誤)

---

### Q8: 如何更改管理員密碼?

```bash
# 方法 1: 透過 Helm 升級
helm upgrade domain-manager ./chart \
  --namespace domain-manager \
  --set admin.password="new-password"

# 方法 2: 手動更新資料庫 (進階)
kubectl exec -it -n domain-manager deployment/domain-manager -- \
  sqlite3 /data/database.db "UPDATE admin_accounts SET password_hash = '...' WHERE username = 'admin';"
```

**建議**: 使用 Helm 升級方式更安全。

---

### Q9: 系統支援多少個域名?

理論上無限制,但建議:
- **小規模**: ≤50 個域名 (推薦配置)
- **中等規模**: 50-100 個域名 (需要適度增加資源)
- **大規模**: >100 個域名 (建議增加資源並監控效能)

如需管理 100+ 個域名,建議調整資源:
```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 200m
    memory: 256Mi
```

---

### Q10: 可以使用萬用字元憑證嗎?

可以,但需要使用 DNS-01 挑戰 (較複雜):

1. 確保 cert-manager 已安裝
2. 配置 DNS 供應商的 API 憑證
3. 建立使用 DNS-01 的 ClusterIssuer
4. 在域名配置中指定使用該 Issuer

**範例**: `*.example.com` 萬用字元憑證可涵蓋所有子域名。

---

## 疑難排解

### 檢查清單

當遇到問題時,請依序檢查:

1. **Kubernetes 叢集健康狀態**
   ```bash
   kubectl get nodes
   kubectl get pods --all-namespaces
   ```

2. **域名管理器 Pod 日誌**
   ```bash
   kubectl logs -n domain-manager deployment/domain-manager --tail=100
   ```

3. **Ingress 控制器狀態**
   ```bash
   kubectl get pods -n ingress-nginx
   kubectl logs -n ingress-nginx deployment/nginx-ingress-controller
   ```

4. **cert-manager 狀態** (如使用)
   ```bash
   kubectl get pods -n cert-manager
   kubectl describe certificate -n domain-manager
   ```

5. **DNS 解析**
   ```bash
   nslookup example.com
   dig example.com
   ```

6. **系統健康檢查** (透過管理介面)
   - 前往 **診斷 > 系統健康**
   - 查看各元件狀態

---

## 下一步

恭喜!您已成功完成快速入門。接下來您可以:

1. **探索 REST API**
   - 閱讀 API 文件: `/docs/api-rest.yaml`
   - 使用 Postman 或 curl 測試 API

2. **整合 MCP (Model Context Protocol)**
   - 配置 Claude Desktop
   - 透過 AI 助理管理域名

3. **設定自動化工作流程**
   - 使用 CI/CD 整合 API
   - 自動化域名配置

4. **監控和維護**
   - 定期檢查憑證狀態
   - 建立備份排程

5. **進階配置**
   - 自訂 Ingress 註解
   - 配置速率限制
   - 整合外部 DNS 供應商

---

## 取得幫助

- **文件**: https://github.com/yourusername/domain-manager/docs
- **問題回報**: https://github.com/yourusername/domain-manager/issues
- **社群論壇**: https://community.example.com

---

## 附錄: 完整範例配置

### values.yaml (完整版本)

```yaml
# 管理員設定
admin:
  username: admin
  password: "change-this-password"
  email: "admin@example.com"

# 映像設定
image:
  repository: domain-manager/domain-manager
  tag: "1.0.0"
  pullPolicy: IfNotPresent

# 服務設定
service:
  type: ClusterIP
  port: 8080

# Ingress 設定
ingress:
  enabled: false  # 透過系統自身管理
  className: "nginx"
  certManager:
    enabled: true
    issuerEmail: "admin@example.com"
    acmeServer: "https://acme-v02.api.letsencrypt.org/directory"

# 持久化設定
persistence:
  enabled: true
  size: 1Gi
  storageClass: ""  # 使用預設 StorageClass
  accessMode: ReadWriteOnce

# 資源限制
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi

# MCP 伺服器
mcp:
  enabled: true
  port: 8080

# 備份設定
backup:
  enabled: true
  schedule: "0 2 * * *"  # 每天凌晨 2 點
  retentionDays: 30

# RBAC 設定
rbac:
  create: true

# ServiceAccount
serviceAccount:
  create: true
  name: ""

# 安全設定
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000

# 健康檢查
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5

# 環境變數
env:
  - name: LOG_LEVEL
    value: "info"
  - name: DATABASE_PATH
    value: "/data/database.db"
  # Let's Encrypt 配置
  - name: LETSENCRYPT_EMAIL
    value: "admin@example.com"
  - name: LETSENCRYPT_ACCOUNT_PATH
    value: "/data/letsencrypt"
  - name: LETSENCRYPT_STAGING
    value: "false"  # 測試環境請設為 "true"
```

---

**文件版本**: 1.0
**最後更新**: 2025-11-07
**預估完成時間**: 15-20 分鐘 (含 DNS 傳播和憑證申請)
