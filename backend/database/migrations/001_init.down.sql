-- Rollback script for initial database schema

-- Drop tables in reverse order to respect foreign key constraints
DROP TABLE IF EXISTS diagnostic_logs;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS admin_accounts;
DROP TABLE IF EXISTS certificates;
DROP TABLE IF EXISTS domains;
