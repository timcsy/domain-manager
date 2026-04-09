# Feature Specification: 多 Ingress Controller 支援

**Feature Branch**: `002-multi-ingress`  
**Created**: 2026-04-09  
**Status**: Draft  
**Input**: User description: "Support multiple Ingress Controllers including K3s Traefik"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 從系統設定切換 Ingress Controller (Priority: P1)

DevOps 工程師在 K3s 叢集上部署 domain-manager，叢集使用 Traefik 作為預設 Ingress Controller。工程師進入系統設定頁面，將預設 Ingress class 從 `nginx` 改為 `traefik`。之後建立的所有域名都會自動使用 Traefik Ingress class。

**Why this priority**: 這是最基本的需求——讓使用者能選擇 Ingress Controller，否則在非 Nginx 環境下完全無法使用。

**Independent Test**: 修改系統設定中的 `default_ingress_class` 為 `traefik`，建立一個新域名，驗證產生的 Ingress 資源的 `ingressClassName` 為 `traefik`。

**Acceptance Scenarios**:

1. **Given** 系統設定中 `default_ingress_class` 為 `traefik`，**When** 使用者建立新域名，**Then** 產生的 Ingress 資源 `spec.ingressClassName` 為 `traefik`
2. **Given** 系統設定中 `default_ingress_class` 為 `nginx`，**When** 使用者建立新域名，**Then** 產生的 Ingress 資源 `spec.ingressClassName` 為 `nginx`
3. **Given** 系統設定中 `default_ingress_class` 從 `nginx` 改為 `traefik`，**When** 使用者更新既有域名，**Then** 更新後的 Ingress 資源使用 `traefik`

---

### User Story 2 - 不同 Controller 的 Annotation 自動適配 (Priority: P2)

DevOps 工程師在 Traefik 環境下建立域名，系統自動產生符合 Traefik 規範的 Ingress annotation（例如 `traefik.ingress.kubernetes.io/router.tls=true`），而非 Nginx 特有的 annotation。

**Why this priority**: 沒有正確的 annotation，Ingress 資源雖然能建立但可能無法正常運作。

**Independent Test**: 分別設定 `traefik` 和 `nginx` 為 Ingress class，建立域名後檢查 Ingress annotation 是否符合各自 Controller 的規範。

**Acceptance Scenarios**:

1. **Given** Ingress class 為 `nginx`，**When** 建立帶 SSL 的域名，**Then** Ingress 包含 Nginx 相關的 TLS annotation
2. **Given** Ingress class 為 `traefik`，**When** 建立帶 SSL 的域名，**Then** Ingress 包含 Traefik 相關的 TLS annotation
3. **Given** 使用者在系統設定中自訂了額外的 Ingress annotation（JSON），**When** 建立域名，**Then** 自訂 annotation 會合併到 Ingress 資源中

---

### User Story 3 - Helm 部署時指定 Ingress Controller (Priority: P3)

DevOps 工程師在 Helm 安裝時透過 `values.yaml` 設定 Ingress Controller 類型，應用程式啟動後自動使用指定的 Controller，無需手動進入 UI 修改設定。

**Why this priority**: 首次部署的便利性，但使用者也可以部署後再從 UI 修改，因此優先級較低。

**Independent Test**: 使用 `--set subdomains.defaultIngressClass=traefik` 安裝 Helm chart，驗證應用程式啟動後系統設定中 `default_ingress_class` 為 `traefik`。

**Acceptance Scenarios**:

1. **Given** Helm values 中 `subdomains.defaultIngressClass` 設為 `traefik`，**When** 首次安裝並啟動，**Then** 系統設定中 `default_ingress_class` 為 `traefik`
2. **Given** Helm values 中 `subdomains.defaultIngressClass` 未設定，**When** 首次安裝，**Then** 系統設定中 `default_ingress_class` 維持資料庫預設值 `nginx`

---

### Edge Cases

- 使用者設定了不存在的 Ingress class（例如 `haproxy`）時，系統仍然建立 Ingress 資源，但在診斷記錄中產生警告
- 既有使用 `nginx` class 建立的 Ingress 資源，切換設定為 `traefik` 後不自動更新，僅影響新建立和手動更新的域名
- 叢集中同時有 Nginx 和 Traefik 時，所有域名使用同一個預設 Ingress class（不支援 per-domain 設定）

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系統建立或更新 Ingress 資源時，MUST 從系統設定讀取 `default_ingress_class`，不得硬編碼
- **FR-002**: 系統 MUST 根據 Ingress class 類型自動套用對應的 annotation 集合
- **FR-003**: 系統 MUST 支援至少 `nginx` 和 `traefik` 兩種 Ingress Controller
- **FR-004**: 使用者 MUST 能透過系統設定頁面修改預設 Ingress class
- **FR-005**: 使用者在系統設定中自訂的 Ingress annotation（JSON 格式）MUST 被合併到產生的 Ingress 資源中
- **FR-006**: Helm values 中的 Ingress class 設定 MUST 在首次部署時正確傳入應用程式
- **FR-007**: 設定不支援的 Ingress class 時，系統 MUST 仍能建立 Ingress 資源，並在診斷記錄中產生警告

### Key Entities

- **Ingress Controller Profile**: 代表一種 Ingress Controller 的配置，包含名稱（nginx/traefik）、預設 annotation 集合、TLS 相關 annotation
- **系統設定 `default_ingress_class`**: 全域預設的 Ingress class 名稱
- **系統設定 `ingress_annotations`**: 使用者自訂的額外 annotation（JSON 格式，已存在於資料庫 schema）

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 程式碼中不存在硬編碼的 Ingress class 字串
- **SC-002**: 使用者從系統設定修改 Ingress class 後，新建立的域名在 5 秒內產生正確 Ingress class 的資源
- **SC-003**: Traefik 和 Nginx 環境下建立的 Ingress 資源各自帶有正確的 annotation
- **SC-004**: 從 Nginx 切換到 Traefik（或反向）不需要重啟應用程式

## Assumptions

- 僅使用標準 Kubernetes Ingress 資源，不使用 Traefik CRD（IngressRoute）
- 不自動偵測叢集中安裝的 Ingress Controller 類型
- 切換 Ingress class 不自動遷移既有的 Ingress 資源
- 所有域名共用同一個 Ingress class（不支援 per-domain 設定）
