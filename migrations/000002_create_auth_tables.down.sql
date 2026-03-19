-- Rollback migration: Drop users and audit_logs tables
-- Warning: This will delete all user accounts and audit trail data!

DROP INDEX IF EXISTS idx_audit_logs_created_at;
DROP INDEX IF EXISTS idx_audit_logs_table_record;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
DROP TABLE IF EXISTS audit_logs;

DROP INDEX IF EXISTS idx_users_username;
DROP TABLE IF EXISTS users;
