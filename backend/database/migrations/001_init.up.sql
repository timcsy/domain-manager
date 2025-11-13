-- Enable foreign key constraints
PRAGMA foreign_keys = ON;

-- Table: domains (域名配置)
CREATE TABLE IF NOT EXISTS domains (
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

CREATE INDEX IF NOT EXISTS idx_domains_domain_name ON domains(domain_name);
CREATE INDEX IF NOT EXISTS idx_domains_status ON domains(status);
CREATE INDEX IF NOT EXISTS idx_domains_enabled ON domains(enabled);
CREATE INDEX IF NOT EXISTS idx_domains_target_service ON domains(target_service, target_namespace);

-- Table: certificates (SSL 憑證)
CREATE TABLE IF NOT EXISTS certificates (
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

CREATE INDEX IF NOT EXISTS idx_certificates_domain_name ON certificates(domain_name);
CREATE INDEX IF NOT EXISTS idx_certificates_status ON certificates(status);
CREATE INDEX IF NOT EXISTS idx_certificates_valid_until ON certificates(valid_until);
CREATE INDEX IF NOT EXISTS idx_certificates_auto_renew ON certificates(auto_renew);
CREATE INDEX IF NOT EXISTS idx_certificates_k8s_secret ON certificates(k8s_secret_name, k8s_secret_namespace);

-- Table: diagnostic_logs (診斷記錄)
CREATE TABLE IF NOT EXISTS diagnostic_logs (
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

CREATE INDEX IF NOT EXISTS idx_diagnostic_logs_domain_id ON diagnostic_logs(domain_id);
CREATE INDEX IF NOT EXISTS idx_diagnostic_logs_log_type ON diagnostic_logs(log_type);
CREATE INDEX IF NOT EXISTS idx_diagnostic_logs_category ON diagnostic_logs(category);
CREATE INDEX IF NOT EXISTS idx_diagnostic_logs_resolved ON diagnostic_logs(resolved);
CREATE INDEX IF NOT EXISTS idx_diagnostic_logs_created_at ON diagnostic_logs(created_at);

-- Table: admin_accounts (管理員帳戶)
CREATE TABLE IF NOT EXISTS admin_accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL, -- bcrypt hash
    email VARCHAR(255) NULL,
    last_login_at TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_accounts_username ON admin_accounts(username);

-- Table: api_keys (API 金鑰)
CREATE TABLE IF NOT EXISTS api_keys (
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

CREATE UNIQUE INDEX IF NOT EXISTS idx_api_keys_key_value ON api_keys(key_value);
CREATE INDEX IF NOT EXISTS idx_api_keys_admin_id ON api_keys(admin_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_enabled ON api_keys(enabled);

-- Table: system_settings (系統設定)
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Insert default system settings
INSERT OR IGNORE INTO system_settings (key, value, description) VALUES
    ('letsencrypt_email', '', 'Let''s Encrypt 註冊郵箱'),
    ('letsencrypt_server', 'https://acme-v02.api.letsencrypt.org/directory', 'ACME 伺服器地址'),
    ('default_ingress_class', 'nginx', '預設 Ingress Class'),
    ('cert_renewal_days', '30', '憑證到期前幾天開始續約'),
    ('cert_manager_enabled', '1', '是否啟用 cert-manager 整合'),
    ('ingress_annotations', '{}', 'Ingress 預設註解 (JSON)'),
    ('backup_enabled', '1', '是否啟用自動備份'),
    ('backup_retention_days', '30', '備份保留天數');

-- Insert default admin account (password: admin - should be changed on first login)
-- Password hash for 'admin' using bcrypt cost 10
INSERT OR IGNORE INTO admin_accounts (username, password_hash, email) VALUES
    ('admin', '$2a$10$qcgCDlh6uLPMEMMKjEzMLew/.4oJvkhdcb21u3diANtVUafhiFdYC', 'admin@localhost');
