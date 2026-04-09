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

待完成：
- [ ] 階段 6：程式化 API / MCP 操作
- [ ] 階段 7：收尾與跨功能
- [ ] Cloudflare API 整合（DNS-01 challenge for wildcard 憑證）

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

- [ ] 完成

**成功標準：**
- [ ] API Key 管理機制
- [ ] 備份功能
- [ ] MCP Server 實作
- [ ] API 文件與測試
- [ ] Helm 與前端增強
- [ ] 整合功能

### 階段 7：收尾與跨功能

- [ ] 完成

**成功標準：**
- [ ] README 與 Quickstart 文件
- [ ] 故障排除指南
- [ ] 架構文件
- [ ] 端對端驗證
- [ ] 程式碼審查與重構

### 未來：Cloudflare 整合

- [ ] 完成

**成功標準：**
- [ ] 透過 Cloudflare API 自動設定 DNS 記錄
- [ ] 支援 DNS-01 challenge 實現 wildcard 憑證
