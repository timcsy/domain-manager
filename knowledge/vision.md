# 願景

## 問題陳述

DevOps 工程師在 Kubernetes 叢集中管理 domain 和 SSL 憑證的流程不直觀：
- 要手動規劃並輸入 K8s 服務名稱，無法從既有服務中直接選取
- cert-manager 設定繁瑣，尤其是 wildcard 憑證搭配 DNS-01 challenge 的場景
- 缺乏圖形化介面，所有操作都得透過 YAML 和 kubectl

## 核心想法

讓子網域快速方便對應到 K8s 服務——透過 Web UI 用下拉式選單選擇服務，搭配 Cloudflare API 和 Let's Encrypt 自動處理 wildcard SSL 憑證。

## 現狀

已完成：
- Go backend REST API（Chi + SQLite）
- Web UI（HTMX + TailwindCSS + Alpine.js）
- K8s Ingress 資源管理
- Let's Encrypt 憑證自動化（lego）
- 子網域驗證、批次操作
- 診斷工具、排程器
- Helm chart 與 Kustomize 部署
- CI/CD 與容器化
- API Key 認證機制
- MCP Server（JSON-RPC 2.0，12 個 Tools + 4 個 Resources）
- 資料庫備份與還原
- 速率限制與請求追蹤
- 系統設定管理介面、API 文件頁面
- 完整文件（quickstart、troubleshooting、architecture）

待完成：
- [ ] Cloudflare API 整合（DNS-01 challenge for wildcard 憑證）
- [ ] 端對端驗證（需在實際 K8s 叢集執行）

## 架構

- **Backend**：Go + Chi framework，SQLite 作為本地狀態儲存
- **Frontend**：HTMX + TailwindCSS + Alpine.js，輕量且不需要複雜的前端建置
- **K8s 整合**：client-go 直接操作 Ingress 資源
- **憑證**：lego 庫處理 Let's Encrypt ACME 流程
- **部署**：Helm chart / Kustomize，容器化單一映像

## 路線圖

### 階段 1-5：核心功能與子網域管理

- [x] 完成

**成功標準：**
- [x] 專案架構與基礎設施建立
- [x] Helm 快速部署
- [x] 外部 Domain 與自動 SSL 憑證
- [x] CI/CD 與容器化
- [x] 子網域管理

### 階段 6：程式化 API / MCP 操作

- [x] 完成

**成功標準：**
- [x] API Key 管理機制
- [x] 備份功能
- [x] MCP Server 實作
- [x] API 文件與測試
- [x] Helm 與前端增強
- [x] 整合功能

### 階段 7：收尾與跨功能

- [x] 完成

**成功標準：**
- [x] README 與 Quickstart 文件
- [x] 故障排除指南
- [x] 架構文件
- [ ] 端對端驗證（需在實際 K8s 叢集執行）
- [x] 程式碼審查與重構

### 未來：多 Ingress Controller 支援（含 K3s Traefik）

- [x] 完成

**成功標準：**
- [x] Ingress class 從系統設定動態讀取，不再硬編碼 nginx
- [x] 支援 Traefik（K3s 預設）和 Nginx Ingress Controller
- [x] 不同 Ingress Controller 的 annotation 差異自動處理

### 未來：Cloudflare DNS + cert-manager 整合

- [x] 完成

**方向**：使用 cert-manager 的 DNS-01 solver 搭配 Cloudflare 免費 DNS API 申請 Let's Encrypt wildcard 憑證。全免費架構（Cloudflare 免費 DNS + Let's Encrypt + cert-manager），不需要公開 80 port 做 HTTP-01 challenge。

**成功標準：**
- [x] 使用者在 UI 設定 Cloudflare API Token
- [x] 系統自動建立 K8s Secret 和 ClusterIssuer（DNS-01 + Cloudflare solver）
- [x] cert-manager 透過 Cloudflare DNS API 完成 DNS-01 challenge
- [x] 成功申請 wildcard 憑證（`*.example.com`）
- [x] 憑證自動續期由 cert-manager 處理
