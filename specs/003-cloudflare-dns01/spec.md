# Feature Specification: Cloudflare DNS + cert-manager DNS-01 整合

**Feature Branch**: `003-cloudflare-dns01`  
**Created**: 2026-04-09  
**Status**: Draft  
**Input**: User description: "Cloudflare DNS + cert-manager DNS-01 integration for wildcard certificates"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 設定 Cloudflare API Token (Priority: P1)

DevOps 工程師進入系統設定頁面，在 Cloudflare 區塊輸入 API Token。系統驗證 token 有效性後儲存，並自動在叢集中建立對應的 Secret 供 cert-manager 使用。

**Why this priority**: 這是所有 DNS-01 功能的前提——沒有有效的 Cloudflare token，後續功能都無法運作。

**Independent Test**: 在 UI 輸入 Cloudflare API Token，確認系統成功驗證並儲存，叢集中出現對應的 Secret。

**Acceptance Scenarios**:

1. **Given** 使用者有 Cloudflare 帳號和有效的 API Token，**When** 在系統設定頁面輸入 token 並儲存，**Then** 系統驗證 token 有效後儲存，並顯示驗證成功狀態
2. **Given** 使用者輸入無效的 API Token，**When** 儲存設定，**Then** 系統顯示 token 無效的錯誤訊息，不儲存
3. **Given** token 已成功儲存，**When** 查看叢集資源，**Then** 對應的 Secret 已建立在正確的 namespace

---

### User Story 2 - 自動建立 DNS-01 ClusterIssuer (Priority: P2)

系統在 Cloudflare token 設定完成後，自動建立或更新 ClusterIssuer，使用 DNS-01 solver 搭配 Cloudflare。cert-manager 使用此 issuer 申請憑證時，會自動透過 Cloudflare DNS API 完成 challenge。

**Why this priority**: ClusterIssuer 是 cert-manager 申請憑證的必要配置，是連接 Cloudflare token 和憑證申請的橋樑。

**Independent Test**: 設定 Cloudflare token 後，確認叢集中出現使用 DNS-01 solver 的 ClusterIssuer。

**Acceptance Scenarios**:

1. **Given** Cloudflare token 已設定，**When** 系統建立 ClusterIssuer，**Then** ClusterIssuer 使用 DNS-01 solver 且參照正確的 Cloudflare token Secret
2. **Given** 使用者更新 Cloudflare token，**When** 系統更新設定，**Then** Secret 內容更新，ClusterIssuer 無需重建（參照同一 Secret 名稱）
3. **Given** 使用者同時需要 HTTP-01 和 DNS-01，**When** 系統建立 ClusterIssuer，**Then** 兩種 solver 並存，DNS-01 用於 wildcard 域名，HTTP-01 用於一般域名

---

### User Story 3 - 申請 Wildcard 憑證 (Priority: P3)

DevOps 工程師在建立域名時選擇使用 wildcard 憑證（例如 `*.example.com`）。系統自動透過 cert-manager + DNS-01 challenge 申請 wildcard 憑證，所有匹配的子網域共用這張憑證。

**Why this priority**: 這是最終使用者價值的交付——能實際用 wildcard 憑證保護子網域。

**Independent Test**: 建立 wildcard 域名配置，確認 cert-manager 透過 DNS-01 成功申請憑證，子網域可透過 HTTPS 存取。

**Acceptance Scenarios**:

1. **Given** Cloudflare token 和 DNS-01 ClusterIssuer 已配置，**When** 使用者建立域名並選擇 wildcard SSL，**Then** cert-manager 透過 DNS-01 challenge 成功申請 `*.example.com` 憑證
2. **Given** wildcard 憑證已申請成功，**When** 使用者建立子網域（如 `app.example.com`），**Then** 子網域自動使用既有的 wildcard 憑證，無需額外申請
3. **Given** wildcard 憑證即將到期，**When** cert-manager 自動續期，**Then** 續期透過 DNS-01 完成，無需人工介入

---

### User Story 4 - Helm 部署時預設 Cloudflare 配置 (Priority: P4)

DevOps 工程師在 Helm 安裝時透過 values.yaml 預設 Cloudflare API Token，首次部署即自動完成所有 DNS-01 配置，無需事後手動設定。

**Why this priority**: 方便首次部署，但使用者也可以部署後再從 UI 設定。

**Independent Test**: 使用 `--set cloudflare.apiToken=xxx` 安裝 Helm chart，確認 Secret 和 ClusterIssuer 自動建立。

**Acceptance Scenarios**:

1. **Given** Helm values 中設定了 Cloudflare API Token，**When** 首次安裝，**Then** Secret、ClusterIssuer（含 DNS-01 solver）自動建立
2. **Given** Helm values 中未設定 Cloudflare Token，**When** 首次安裝，**Then** 系統使用 HTTP-01 solver 作為預設，使用者可稍後在 UI 設定 Cloudflare

---

### Edge Cases

- Cloudflare API Token 被撤銷後，cert-manager 續期失敗 → 系統在診斷記錄中警告，提示使用者更新 token
- 使用者的 Cloudflare 帳號免費方案僅有部分 zone → token 需有對應 zone 的 DNS 編輯權限
- cert-manager 未安裝 → 系統在設定頁面顯示 cert-manager 未偵測到的警告
- 已有使用 HTTP-01 的 ClusterIssuer → 升級為 DNS-01 時不中斷既有憑證

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 使用者 MUST 能在系統設定頁面輸入和管理 Cloudflare API Token
- **FR-002**: 系統 MUST 在儲存 token 前驗證其有效性
- **FR-003**: 系統 MUST 將 Cloudflare token 安全儲存為叢集中的 Secret 資源
- **FR-004**: 系統 MUST 自動建立或更新使用 DNS-01 solver 的 ClusterIssuer
- **FR-005**: ClusterIssuer MUST 同時支援 DNS-01（wildcard）和 HTTP-01（一般域名）solver
- **FR-006**: 使用者建立域名時 MUST 能選擇使用 wildcard 憑證
- **FR-007**: 選擇 wildcard 憑證時，系統 MUST 使用 DNS-01 ClusterIssuer 申請 `*.rootdomain` 憑證
- **FR-008**: 同一 root domain 下的子網域 MUST 能共用已申請的 wildcard 憑證
- **FR-009**: Helm values MUST 支援預設 Cloudflare API Token 配置
- **FR-010**: Cloudflare token 失效時，系統 MUST 在診斷記錄中產生警告

### Key Entities

- **Cloudflare API Token**: 使用者提供的 Cloudflare API token，用於 DNS-01 challenge。存放在系統設定和 K8s Secret 中
- **ClusterIssuer (DNS-01)**: cert-manager 資源，配置 Cloudflare DNS solver
- **Wildcard Certificate**: 由 cert-manager 透過 DNS-01 申請的 `*.example.com` 格式憑證

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 使用者從設定 Cloudflare token 到成功申請 wildcard 憑證的整個流程在 5 分鐘內完成（不含 DNS 傳播時間）
- **SC-002**: wildcard 憑證成功申請後，新增子網域 HTTPS 即時可用（10 秒內）
- **SC-003**: 憑證到期前自動續期成功率達 100%（在 token 有效的前提下）
- **SC-004**: 整個方案零成本運作（Cloudflare 免費 DNS + Let's Encrypt 免費憑證）

## Assumptions

- 使用者已在 Cloudflare 免費方案中設定好 DNS zone
- Cloudflare API Token 具有對應 zone 的 DNS 編輯權限（Zone:DNS:Edit）
- K8s 叢集已安裝 cert-manager（Helm chart 預設包含）
- 不自動建立或管理 Cloudflare DNS zone 和 A 記錄
- DNS-01 和 HTTP-01 solver 共存於同一個 ClusterIssuer
