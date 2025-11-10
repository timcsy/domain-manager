# 資料模型設計: Kubernetes 域名管理器

**功能分支**: `001-k8s-domain-manager`
**建立日期**: 2025-11-07
**狀態**: Draft
**版本**: 1.0

---

## 1. 資料庫 Schema (SQLite)

### 1.1 domains (域名配置)

儲存所有域名和子域名的配置資訊。

```sql
CREATE TABLE domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_name VARCHAR(255) NOT NULL UNIQUE,
    target_service VARCHAR(255) NOT NULL,
    target_namespace VARCHAR(255) NOT NULL DEFAULT 'default',
    target_port INTEGER NOT NULL,
    ssl_mode VARCHAR(20) NOT NULL DEFAULT 'auto', -- 'auto' (Let's Encrypt) or 'manual' (使用者上傳)
    certificate_id INTEGER NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'active', 'error', 'deleted'
    enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (certificate_id) REFERENCES certificates(id) ON DELETE SET NULL
);

CREATE INDEX idx_domains_domain_name ON domains(domain_name);
CREATE INDEX idx_domains_status ON domains(status);
CREATE INDEX idx_domains_enabled ON domains(enabled);
CREATE INDEX idx_domains_target_service ON domains(target_service, target_namespace);
```

**欄位說明**:
- `id`: 主鑰,自動遞增
- `domain_name`: 域名或子域名 (如 example.com, api.example.com),必須唯一
- `target_service`: 目標 Kubernetes Service 名稱
- `target_namespace`: 目標 Service 所在的命名空間
- `target_port`: 目標 Service 的連接埠號
- `ssl_mode`: SSL 憑證模式 ('auto' 使用 Let's Encrypt, 'manual' 使用者上傳)
- `certificate_id`: 關聯的憑證 ID (外鍵,可為 NULL)
- `status`: 域名狀態 (pending 待配置, active 運行中, error 錯誤, deleted 已刪除)
- `enabled`: 是否啟用 (軟刪除標記)
- `created_at`: 建立時間
- `updated_at`: 最後更新時間

---

### 1.2 certificates (SSL 憑證)

儲存 SSL/TLS 憑證資訊和內容。

```sql
CREATE TABLE certificates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_name VARCHAR(255) NOT NULL,
    source VARCHAR(20) NOT NULL DEFAULT 'letsencrypt', -- 'letsencrypt' or 'manual'
    certificate_pem TEXT NOT NULL,
    private_key_pem TEXT NOT NULL,
    issuer VARCHAR(255) NULL,
    valid_from TIMESTAMP NOT NULL,
    valid_until TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'valid', -- 'valid', 'expiring', 'expired', 'revoked'
    k8s_secret_name VARCHAR(255) NOT NULL,
    k8s_secret_namespace VARCHAR(255) NOT NULL DEFAULT 'default',
    auto_renew BOOLEAN NOT NULL DEFAULT 1,
    last_renewal_attempt TIMESTAMP NULL,
    renewal_error TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_certificates_domain_name ON certificates(domain_name);
CREATE INDEX idx_certificates_status ON certificates(status);
CREATE INDEX idx_certificates_valid_until ON certificates(valid_until);
CREATE INDEX idx_certificates_auto_renew ON certificates(auto_renew);
CREATE INDEX idx_certificates_k8s_secret ON certificates(k8s_secret_name, k8s_secret_namespace);
```

**欄位說明**:
- `id`: 主鑰,自動遞增
- `domain_name`: 憑證適用的域名
- `source`: 憑證來源 ('letsencrypt' 自動申請, 'manual' 使用者上傳)
- `certificate_pem`: 憑證內容 (PEM 格式)
- `private_key_pem`: 私鑰內容 (PEM 格式,加密儲存)
- `issuer`: 憑證頒發者
- `valid_from`: 憑證生效時間
- `valid_until`: 憑證到期時間
- `status`: 憑證狀態 (valid 有效, expiring 即將到期 [30天內], expired 已過期, revoked 已撤銷)
- `k8s_secret_name`: Kubernetes Secret 名稱
- `k8s_secret_namespace`: Kubernetes Secret 命名空間
- `auto_renew`: 是否自動續約 (僅 Let's Encrypt 憑證)
- `last_renewal_attempt`: 最後續約嘗試時間
- `renewal_error`: 續約錯誤訊息
- `created_at`: 建立時間
- `updated_at`: 最後更新時間

---

### 1.3 diagnostic_logs (診斷記錄)

記錄域名配置和憑證管理過程中的問題和警告。

```sql
CREATE TABLE diagnostic_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_id INTEGER NULL,
    log_type VARCHAR(20) NOT NULL, -- 'info', 'warning', 'error'
    category VARCHAR(50) NOT NULL, -- 'dns', 'certificate', 'ingress', 'service', 'system'
    message TEXT NOT NULL,
    details TEXT NULL, -- JSON 格式的詳細資訊
    resolved BOOLEAN NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE
);

CREATE INDEX idx_diagnostic_logs_domain_id ON diagnostic_logs(domain_id);
CREATE INDEX idx_diagnostic_logs_log_type ON diagnostic_logs(log_type);
CREATE INDEX idx_diagnostic_logs_category ON diagnostic_logs(category);
CREATE INDEX idx_diagnostic_logs_resolved ON diagnostic_logs(resolved);
CREATE INDEX idx_diagnostic_logs_created_at ON diagnostic_logs(created_at);
```

**欄位說明**:
- `id`: 主鑰,自動遞增
- `domain_id`: 關聯的域名 ID (外鍵,可為 NULL 表示系統級日誌)
- `log_type`: 日誌類型 (info 資訊, warning 警告, error 錯誤)
- `category`: 日誌分類 (dns DNS配置, certificate 憑證, ingress Ingress配置, service 服務, system 系統)
- `message`: 日誌訊息
- `details`: 詳細資訊 (JSON 格式,如錯誤堆疊、診斷結果)
- `resolved`: 是否已解決
- `created_at`: 建立時間

---

### 1.4 admin_accounts (管理員帳戶)

儲存管理員的身份驗證資訊。

```sql
CREATE TABLE admin_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL, -- bcrypt hash
    email VARCHAR(255) NULL,
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_admin_accounts_username ON admin_accounts(username);
```

**欄位說明**:
- `id`: 主鑰,自動遞增
- `username`: 使用者名稱,唯一
- `password_hash`: bcrypt 加密的密碼雜湊值
- `email`: 電子郵件地址 (選填,用於 Let's Encrypt 通知)
- `last_login_at`: 最後登入時間
- `created_at`: 建立時間
- `updated_at`: 最後更新時間

---

### 1.5 api_keys (API 金鑰)

儲存用於 REST API 和 MCP 身份驗證的金鑰。

```sql
CREATE TABLE api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key_value VARCHAR(64) NOT NULL UNIQUE, -- SHA-256 hash of the actual key
    key_name VARCHAR(100) NOT NULL,
    admin_id INTEGER NOT NULL,
    permissions TEXT NOT NULL DEFAULT '["read"]', -- JSON array: ["read", "write", "delete"]
    enabled BOOLEAN NOT NULL DEFAULT 1,
    last_used_at TIMESTAMP NULL,
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (admin_id) REFERENCES admin_accounts(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_api_keys_key_value ON api_keys(key_value);
CREATE INDEX idx_api_keys_admin_id ON api_keys(admin_id);
CREATE INDEX idx_api_keys_enabled ON api_keys(enabled);
```

**欄位說明**:
- `id`: 主鑰,自動遞增
- `key_value`: API 金鑰的 SHA-256 雜湊值
- `key_name`: 金鑰名稱 (便於識別,如 "CI/CD Pipeline Key")
- `admin_id`: 關聯的管理員 ID (外鍵)
- `permissions`: 權限列表 (JSON 陣列,read 讀取, write 寫入, delete 刪除)
- `enabled`: 是否啟用
- `last_used_at`: 最後使用時間
- `expires_at`: 到期時間 (可為 NULL 表示永不過期)
- `created_at`: 建立時間

---

### 1.6 system_settings (系統設定)

儲存全域系統設定。

```sql
CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 預設系統設定
INSERT INTO system_settings (key, value, description) VALUES
    ('letsencrypt_email', '', 'Let''s Encrypt 註冊郵箱'),
    ('letsencrypt_server', 'https://acme-v02.api.letsencrypt.org/directory', 'ACME 伺服器地址'),
    ('default_ingress_class', 'nginx', '預設 Ingress Class'),
    ('cert_renewal_days', '30', '憑證到期前幾天開始續約'),
    ('cert_manager_enabled', '1', '是否啟用 cert-manager 整合'),
    ('ingress_annotations', '{}', 'Ingress 預設註解 (JSON)'),
    ('backup_enabled', '1', '是否啟用自動備份'),
    ('backup_retention_days', '30', '備份保留天數');
```

**欄位說明**:
- `key`: 設定鍵 (主鑰)
- `value`: 設定值 (文字格式,支援 JSON)
- `description`: 設定說明
- `updated_at`: 最後更新時間

---

## 2. 實體關係圖 (ERD)

```
┌─────────────────────┐
│   admin_accounts    │
│─────────────────────│
│ PK: id              │
│ username (UNIQUE)   │
│ password_hash       │
│ email               │
│ last_login_at       │
└──────────┬──────────┘
           │ 1
           │
           │ N
┌──────────▼──────────┐
│      api_keys       │
│─────────────────────│
│ PK: id              │
│ FK: admin_id        │
│ key_value (UNIQUE)  │
│ key_name            │
│ permissions         │
│ enabled             │
│ last_used_at        │
│ expires_at          │
└─────────────────────┘


┌─────────────────────┐
│      domains        │
│─────────────────────│
│ PK: id              │
│ domain_name (UNIQUE)│
│ target_service      │
│ target_namespace    │
│ target_port         │
│ ssl_mode            │
│ FK: certificate_id  │◄─────┐
│ status              │      │
│ enabled             │      │
└──────────┬──────────┘      │
           │ 1                │ N
           │                  │
           │ N                │
┌──────────▼──────────┐       │
│  diagnostic_logs    │       │
│─────────────────────│       │
│ PK: id              │       │
│ FK: domain_id       │       │
│ log_type            │       │
│ category            │       │
│ message             │       │
│ details             │       │
│ resolved            │       │
└─────────────────────┘       │
                              │
                              │
┌─────────────────────────────┘
│
│
│
│   ┌─────────────────────┐
└───│   certificates      │
    │─────────────────────│
    │ PK: id              │
    │ domain_name         │
    │ source              │
    │ certificate_pem     │
    │ private_key_pem     │
    │ issuer              │
    │ valid_from          │
    │ valid_until         │
    │ status              │
    │ k8s_secret_name     │
    │ k8s_secret_namespace│
    │ auto_renew          │
    │ last_renewal_attempt│
    │ renewal_error       │
    └─────────────────────┘


┌─────────────────────┐
│  system_settings    │
│─────────────────────│
│ PK: key             │
│ value               │
│ description         │
└─────────────────────┘
```

### 關係說明

1. **admin_accounts ↔ api_keys** (一對多)
   - 一個管理員可以擁有多個 API 金鑰
   - 刪除管理員時級聯刪除其所有 API 金鑰

2. **domains ↔ certificates** (多對一)
   - 多個域名可以共用同一個憑證 (如萬用字元憑證)
   - 憑證被刪除時,域名的 certificate_id 設為 NULL

3. **domains ↔ diagnostic_logs** (一對多)
   - 一個域名可以有多條診斷記錄
   - 刪除域名時級聯刪除其所有診斷記錄

4. **system_settings** (獨立表)
   - 不與其他表建立外鍵關係
   - 儲存全域配置

---

## 3. 狀態機 (State Machines)

### 3.1 域名狀態 (domains.status)

```
       ┌──────────┐
       │          │
       │ pending  │◄─────── 初始狀態 (新建域名)
       │          │
       └────┬─────┘
            │
            │ Ingress 建立成功 + 憑證申請中
            │
            ▼
       ┌──────────┐
       │          │
       │  active  │◄─────── 域名正常運行
       │          │
       └────┬─────┘
            │
            │ DNS 錯誤、Service 不存在、憑證失效
            │
            ▼
       ┌──────────┐         重試成功
       │          │─────────────────┐
       │  error   │                 │
       │          │◄────────────────┘
       └────┬─────┘
            │
            │ 手動刪除
            │
            ▼
       ┌──────────┐
       │          │
       │ deleted  │ (軟刪除,enabled=0)
       │          │
       └──────────┘
```

**狀態轉換規則**:

| 當前狀態 | 觸發事件 | 下一狀態 | 動作 |
|---------|---------|---------|------|
| `pending` | Ingress 建立成功 | `active` | 更新 updated_at |
| `pending` | Ingress 建立失敗 | `error` | 記錄診斷日誌 |
| `active` | DNS 驗證失敗 | `error` | 記錄診斷日誌 |
| `active` | 服務不存在 | `error` | 記錄診斷日誌 |
| `active` | 憑證過期 | `error` | 記錄診斷日誌 |
| `error` | 問題已修復 | `active` | 清除相關診斷日誌 |
| `error` | 手動刪除 | `deleted` | 設定 enabled=0 |
| `active` | 手動刪除 | `deleted` | 設定 enabled=0 |

---

### 3.2 憑證狀態 (certificates.status)

```
       ┌──────────┐
       │          │
       │  valid   │◄─────── 憑證有效 (距離到期 >30 天)
       │          │
       └────┬─────┘
            │
            │ 到期時間 ≤ 30 天
            │
            ▼
       ┌──────────┐         續約成功
       │          │─────────────────┐
       │ expiring │                 │
       │          │◄────────────────┘
       └────┬─────┘
            │
            │ 到期時間已過 或 續約多次失敗
            │
            ▼
       ┌──────────┐
       │          │
       │ expired  │
       │          │
       └────┬─────┘
            │
            │ 管理員手動撤銷
            │
            ▼
       ┌──────────┐
       │          │
       │ revoked  │
       │          │
       └──────────┘
```

**狀態轉換規則**:

| 當前狀態 | 觸發事件 | 下一狀態 | 動作 |
|---------|---------|---------|------|
| `valid` | 距離到期 ≤ 30 天 | `expiring` | 更新 status, 顯示警告 |
| `expiring` | 自動續約成功 | `valid` | 更新憑證資訊 |
| `expiring` | 已過期 | `expired` | 更新 status, 記錄錯誤 |
| `expired` | 重新申請成功 | `valid` | 建立新憑證記錄 |
| `*` | 手動撤銷 | `revoked` | 更新 status |

---

## 4. 驗證規則 (Validation Rules)

### 4.1 域名驗證

```go
// 域名格式驗證規則
func ValidateDomainName(domain string) error {
    // 1. 長度限制: 1-253 字元
    if len(domain) == 0 || len(domain) > 253 {
        return errors.New("域名長度必須在 1-253 字元之間")
    }

    // 2. 格式驗證: 符合 RFC 1123
    pattern := `^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`
    matched, _ := regexp.MatchString(pattern, domain)
    if !matched {
        return errors.New("域名格式不正確")
    }

    // 3. 禁止使用保留域名
    reserved := []string{"localhost", "example.com", "test.local"}
    for _, r := range reserved {
        if strings.Contains(domain, r) {
            return errors.New("不允許使用保留域名")
        }
    }

    // 4. 標籤長度限制: 每個標籤 ≤ 63 字元
    labels := strings.Split(domain, ".")
    for _, label := range labels {
        if len(label) > 63 {
            return errors.New("域名標籤長度不能超過 63 字元")
        }
    }

    return nil
}
```

### 4.2 憑證驗證

```go
// 憑證格式驗證
func ValidateCertificate(certPEM, keyPEM string) error {
    // 1. 解析憑證
    block, _ := pem.Decode([]byte(certPEM))
    if block == nil || block.Type != "CERTIFICATE" {
        return errors.New("無效的憑證格式")
    }

    cert, err := x509.ParseCertificate(block.Bytes)
    if err != nil {
        return fmt.Errorf("無法解析憑證: %w", err)
    }

    // 2. 檢查憑證是否已過期
    now := time.Now()
    if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
        return errors.New("憑證已過期或尚未生效")
    }

    // 3. 解析私鑰
    keyBlock, _ := pem.Decode([]byte(keyPEM))
    if keyBlock == nil {
        return errors.New("無效的私鑰格式")
    }

    // 4. 驗證私鑰與憑證匹配
    privateKey, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
    if err != nil {
        // 嘗試 PKCS8 格式
        key, err := x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
        if err != nil {
            return errors.New("無法解析私鑰")
        }
        var ok bool
        privateKey, ok = key.(*rsa.PrivateKey)
        if !ok {
            return errors.New("私鑰類型不支援")
        }
    }

    // 驗證公私鑰匹配
    publicKey := cert.PublicKey.(*rsa.PublicKey)
    if publicKey.N.Cmp(privateKey.N) != 0 {
        return errors.New("私鑰與憑證不匹配")
    }

    return nil
}
```

### 4.3 服務驗證

```go
// Kubernetes 服務驗證
func ValidateServiceMapping(service, namespace string, port int) error {
    // 1. 服務名稱格式 (DNS-1035 label)
    pattern := `^[a-z]([-a-z0-9]*[a-z0-9])?$`
    matched, _ := regexp.MatchString(pattern, service)
    if !matched {
        return errors.New("服務名稱格式不正確")
    }

    // 2. 命名空間格式
    matched, _ = regexp.MatchString(pattern, namespace)
    if !matched {
        return errors.New("命名空間格式不正確")
    }

    // 3. 連接埠範圍
    if port < 1 || port > 65535 {
        return errors.New("連接埠必須在 1-65535 之間")
    }

    // 4. 驗證服務是否存在 (查詢 K8s API)
    exists, err := k8sClient.ServiceExists(namespace, service)
    if err != nil {
        return fmt.Errorf("無法驗證服務存在性: %w", err)
    }
    if !exists {
        return errors.New("目標服務不存在")
    }

    return nil
}
```

### 4.4 資料完整性規則

1. **唯一性約束**:
   - `domains.domain_name` 必須唯一
   - `admin_accounts.username` 必須唯一
   - `api_keys.key_value` 必須唯一

2. **外鍵約束**:
   - `domains.certificate_id` → `certificates.id` (SET NULL on delete)
   - `api_keys.admin_id` → `admin_accounts.id` (CASCADE on delete)
   - `diagnostic_logs.domain_id` → `domains.id` (CASCADE on delete)

3. **檢查約束**:
   - `domains.ssl_mode` IN ('auto', 'manual')
   - `domains.status` IN ('pending', 'active', 'error', 'deleted')
   - `certificates.source` IN ('letsencrypt', 'manual')
   - `certificates.status` IN ('valid', 'expiring', 'expired', 'revoked')
   - `diagnostic_logs.log_type` IN ('info', 'warning', 'error')
   - `api_keys.permissions` 必須是有效的 JSON 陣列

4. **業務邏輯約束**:
   - 當 `domains.ssl_mode = 'auto'` 時,`domains.certificate_id` 必須關聯到 `certificates.source = 'letsencrypt'` 的憑證
   - 當 `domains.status = 'deleted'` 時,`domains.enabled` 必須為 `0`
   - 當 `certificates.auto_renew = 1` 時,`certificates.source` 必須為 `'letsencrypt'`

---

## 5. 資料庫遷移 (Migration)

### 初始化腳本 (001_init.sql)

```sql
-- 啟用外鍵約束
PRAGMA foreign_keys = ON;

-- 建立所有表
-- (見上述 Schema 定義)

-- 建立預設管理員帳戶 (密碼: admin, 實際應用時應在安裝時設定)
INSERT INTO admin_accounts (username, password_hash, email) VALUES
    ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '');

-- 建立預設 API 金鑰 (僅用於開發,生產環境應刪除)
-- INSERT INTO api_keys (key_value, key_name, admin_id, permissions) VALUES
--     (sha256('dev-api-key'), 'Development Key', 1, '["read","write","delete"]');
```

### 遷移管理

使用 [golang-migrate](https://github.com/golang-migrate/migrate) 或類似工具管理資料庫遷移:

```
migrations/
├── 001_init.up.sql          # 初始化資料庫
├── 001_init.down.sql        # 回滾初始化
├── 002_add_indexes.up.sql   # 新增索引
├── 002_add_indexes.down.sql # 回滾索引
└── ...
```

---

## 6. 資料存取層 (Repository Pattern)

### 6.1 Domain Repository 介面

```go
type DomainRepository interface {
    Create(domain *Domain) error
    GetByID(id int64) (*Domain, error)
    GetByName(name string) (*Domain, error)
    List(filter DomainFilter) ([]*Domain, error)
    Update(domain *Domain) error
    Delete(id int64) error
    SoftDelete(id int64) error // 軟刪除 (設定 enabled=0)
    Count(filter DomainFilter) (int, error)
}

type DomainFilter struct {
    Status      string
    Enabled     *bool
    ServiceName string
    Namespace   string
    Limit       int
    Offset      int
}
```

### 6.2 Certificate Repository 介面

```go
type CertificateRepository interface {
    Create(cert *Certificate) error
    GetByID(id int64) (*Certificate, error)
    GetByDomain(domain string) (*Certificate, error)
    List(filter CertificateFilter) ([]*Certificate, error)
    Update(cert *Certificate) error
    Delete(id int64) error
    GetExpiring(days int) ([]*Certificate, error) // 取得即將到期的憑證
}

type CertificateFilter struct {
    Status    string
    Source    string
    AutoRenew *bool
    Limit     int
    Offset    int
}
```

---

## 7. 資料安全性

### 7.1 敏感資料加密

- **密碼**: 使用 bcrypt 加密 (cost=10)
- **私鑰**: 使用 AES-256-GCM 加密儲存,密鑰從環境變數或 Kubernetes Secret 讀取
- **API 金鑰**: 儲存 SHA-256 雜湊值,不儲存明文

### 7.2 備份策略

```go
// 自動備份函數
func BackupDatabase(dbPath string) error {
    timestamp := time.Now().Format("20060102_150405")
    backupPath := fmt.Sprintf("%s.backup.%s", dbPath, timestamp)

    // 複製資料庫檔案
    input, err := ioutil.ReadFile(dbPath)
    if err != nil {
        return err
    }

    err = ioutil.WriteFile(backupPath, input, 0644)
    if err != nil {
        return err
    }

    // 壓縮備份
    err = compressFile(backupPath)
    if err != nil {
        return err
    }

    // 清理舊備份 (保留 30 天)
    cleanOldBackups(dbPath, 30)

    return nil
}
```

---

## 8. 效能考量

### 8.1 索引策略

- 為所有外鍵建立索引
- 為常用查詢條件建立索引 (domain_name, status, enabled)
- 為時間欄位建立索引 (created_at, valid_until)

### 8.2 查詢優化

```sql
-- 最佳化查詢: 取得即將到期的活躍域名憑證
SELECT
    d.domain_name,
    c.valid_until,
    c.status,
    c.renewal_error
FROM domains d
JOIN certificates c ON d.certificate_id = c.id
WHERE d.enabled = 1
  AND d.status = 'active'
  AND c.status IN ('valid', 'expiring')
  AND c.valid_until < datetime('now', '+30 days')
ORDER BY c.valid_until ASC;
```

### 8.3 資料庫大小預估

假設管理 100 個域名:

| 表名 | 記錄數 | 單筆大小 | 總大小 |
|-----|-------|---------|--------|
| domains | 100 | ~500 bytes | 50 KB |
| certificates | 100 | ~8 KB (含憑證) | 800 KB |
| diagnostic_logs | 1000 | ~300 bytes | 300 KB |
| admin_accounts | 1 | ~200 bytes | 0.2 KB |
| api_keys | 5 | ~200 bytes | 1 KB |
| system_settings | 10 | ~100 bytes | 1 KB |

**預估總大小**: ~1.2 MB (100 個域名)

---

## 9. 資料模型摘要

### 表結構統計

- **總表數**: 6 個
- **核心業務表**: 2 個 (domains, certificates)
- **支援表**: 4 個 (diagnostic_logs, admin_accounts, api_keys, system_settings)

### 關鍵關係

1. **domains ← certificates**: 多對一 (多個域名可共用一個憑證)
2. **domains → diagnostic_logs**: 一對多 (一個域名有多條日誌)
3. **admin_accounts → api_keys**: 一對多 (一個管理員有多個金鑰)

### 設計決策

1. **選擇 SQLite**: 簡單、零配置、適合小規模管理 (≤100 域名)
2. **軟刪除**: 保留歷史記錄,便於審計
3. **外鍵約束**: 確保資料完整性
4. **狀態機**: 明確定義域名和憑證的生命週期
5. **索引優化**: 為常用查詢建立索引,提升效能

---

**文件版本**: 1.0
**最後更新**: 2025-11-07
