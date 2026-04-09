# Research: Cloudflare DNS + cert-manager DNS-01 整合

## R1: Cloudflare API Token 驗證方式

**Decision**: 使用 Cloudflare API v4 的 `GET /user/tokens/verify` 端點驗證 token 有效性。

**Rationale**: 這是 Cloudflare 官方推薦的驗證方式，只需一次 HTTP 呼叫即可確認 token 是否有效，不需要額外依賴。

**API 呼叫**:
```
GET https://api.cloudflare.com/client/v4/user/tokens/verify
Authorization: Bearer <token>
```
成功回傳 `{"result":{"status":"active"}}`。

**Alternatives considered**:
- 直接嘗試列出 zones：能驗證但可能因權限不足失敗，不代表 token 無效 → 拒絕
- 不驗證，直接儲存：使用者可能等到 challenge 失敗才發現 token 錯誤 → 拒絕

## R2: cert-manager DNS-01 Cloudflare solver 配置

**Decision**: 在 ClusterIssuer 中同時配置 HTTP-01 和 DNS-01 solver，DNS-01 使用 selector 匹配 wildcard 域名。

**Rationale**: cert-manager 支援在同一個 ClusterIssuer 中配置多個 solver，並透過 selector 決定使用哪個。wildcard 域名必須使用 DNS-01，一般域名可繼續使用 HTTP-01。

**ClusterIssuer 配置結構**:
```yaml
solvers:
- dns01:
    cloudflare:
      apiTokenSecretRef:
        name: cloudflare-api-token
        key: api-token
  selector:
    dnsNames:
    - "*.example.com"
- http01:
    ingress:
      class: nginx  # 或 traefik
```

**Alternatives considered**:
- 建立獨立的 ClusterIssuer for DNS-01：增加管理複雜度，使用者需選擇 issuer → 拒絕
- 完全替換為 DNS-01：不需要 wildcard 的域名也走 DNS-01，浪費 Cloudflare API 呼叫 → 拒絕

## R3: K8s Secret 管理策略

**Decision**: 透過 Go client-go 程式碼動態建立/更新 Secret，不依賴 Helm template。

**Rationale**: 使用者在 UI 設定 token 時需要即時建立 Secret，不能等 Helm upgrade。Helm template 的 Secret 作為首次部署時的預設值。

**Secret 格式**:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: cloudflare-api-token
  namespace: cert-manager  # cert-manager 預設 namespace
type: Opaque
stringData:
  api-token: <cloudflare-api-token>
```

**Alternatives considered**:
- 僅靠 Helm template：使用者修改 token 需 helm upgrade → 違反「預設先行」原則
- 存在 domain-manager namespace：cert-manager 可能無法跨 namespace 讀取 → 需確認，但 ClusterIssuer 可指定 Secret namespace

## R4: Wildcard 憑證與 Ingress 的對應方式

**Decision**: 申請 wildcard 憑證（`*.example.com`）後存為 K8s TLS Secret，該 root domain 下所有子網域的 Ingress 共用此 Secret。

**Rationale**: 既有的 `FindApplicableWildcardCertificate` 已實作 wildcard 匹配邏輯。cert-manager 會自動管理 TLS Secret 的建立和續期。

**流程**:
1. 使用者建立 `*.example.com` 的 Certificate 資源
2. cert-manager 透過 DNS-01 申請憑證，存入指定的 TLS Secret
3. 建立子網域（如 `app.example.com`）時，Ingress TLS 指向該 Secret

**Alternatives considered**:
- 每個子網域各申請一張憑證：浪費 Let's Encrypt rate limit，增加管理複雜度 → 拒絕
- 應用層自己管理 wildcard 憑證生命週期：重複 cert-manager 的功能 → 拒絕

## R5: Cloudflare API Token 所需權限

**Decision**: 需要 `Zone:DNS:Edit` 權限，用於 DNS-01 challenge 期間建立和清除 TXT 記錄。

**Rationale**: cert-manager Cloudflare solver 需要能建立 `_acme-challenge.example.com` TXT 記錄。`Zone:DNS:Edit` 是最小必要權限。

**建議使用者建立 token 時的設定**:
- Permissions: Zone - DNS - Edit
- Zone Resources: Include - Specific zone（或 All zones）
