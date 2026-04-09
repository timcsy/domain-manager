# Research: 多 Ingress Controller 支援

## R1: Traefik Ingress Annotation 規範

**Decision**: Traefik 使用標準 Kubernetes Ingress 資源，搭配 `traefik.ingress.kubernetes.io/` 前綴的 annotation。

**Rationale**: Traefik v2+ 完整支援標準 K8s Ingress 資源（除了 IngressRoute CRD）。使用標準 Ingress 可同時相容 Nginx 和 Traefik，無需引入 CRD 依賴。

**Key Traefik Annotations**:
- `traefik.ingress.kubernetes.io/router.tls`: `"true"` — 啟用 TLS
- `traefik.ingress.kubernetes.io/router.entrypoints`: `"websecure"` — 指定 HTTPS entrypoint

**Key Nginx Annotations**:
- `nginx.ingress.kubernetes.io/ssl-redirect`: `"true"` — HTTP 重導向 HTTPS
- `nginx.ingress.kubernetes.io/force-ssl-redirect`: `"true"` — 強制 SSL
- `nginx.ingress.kubernetes.io/proxy-body-size`: 設定 body 大小限制

**Alternatives considered**:
- Traefik IngressRoute CRD: 更強大但增加 CRD 依賴，不相容 Nginx → 拒絕
- 不加 annotation、僅設 ingressClassName: 基本可用但無法控制 TLS 行為 → 拒絕

## R2: 如何從系統設定讀取 Ingress Class

**Decision**: 在 `domain_service.go` 和 `certificate_service.go` 中，每次建立/更新 Ingress 時透過 `settingsService.GetSetting("default_ingress_class")` 讀取。

**Rationale**: 系統設定表已有 `default_ingress_class` 欄位（Phase 2 建立）。`settingsService` 已在 handlers.go 初始化。需要讓 service 層能存取 settings。

**Alternatives considered**:
- 環境變數: 修改需重啟 → 不符合 SC-004（不需重啟）
- 記憶體快取 + 定時刷新: 過度設計 → YAGNI
- 直接查 DB: 每次建 Ingress 查一次 settings 不會有效能問題（SQLite 本地） → 採用

## R3: Ingress Controller Profile 設計

**Decision**: 使用 Go map 定義已知 controller 的預設 annotation，未知 controller 使用空 annotation（僅套用使用者自訂 annotation）。

**Rationale**: 簡單優先。目前只需支援 nginx 和 traefik 兩種，用 map 足夠。未來若需支援更多可擴展。

**Alternatives considered**:
- Interface + 多態: 過度抽象 → YAGNI
- 設定檔驅動: 使用者需理解 annotation 格式 → 違反「預設先行」原則
- Plugin 機制: 嚴重過度設計 → 拒絕

## R4: 既有 Ingress 資源的遷移策略

**Decision**: 不自動遷移。切換 Ingress class 僅影響新建立和手動更新的域名。

**Rationale**: 自動遷移可能導致服務中斷（舊 controller 不再管理、新 controller 需要時間接管）。DevOps 工程師應手動控制遷移節奏。

**Alternatives considered**:
- 批次遷移指令: 有用但增加複雜度，可作為後續功能 → 延後
- 背景自動遷移: 風險高、難回滾 → 拒絕
