-- Migration: Create audit_logs table for authentication compliance
-- This migration depends on users table created in 000001_create_initial_tables.up.sql

-- 1. Tabel Audit Logs (Untuk Compliance & Forensics)
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id),
    action VARCHAR(20) NOT NULL CHECK (action IN ('CREATE', 'UPDATE', 'DELETE')),
    table_name VARCHAR(50) NOT NULL,
    record_id INT NOT NULL,
    old_values TEXT,
    new_values TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_table_record ON audit_logs(table_name, record_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at DESC);

-- ============================================================================
-- Initial data: Create default admin user if not exists
-- ============================================================================
-- PASSWORD FLOW & VERIFICATION:
-- 1. ORIGINAL PASSWORD: admin123 (plaintext - NEVER stored in database)
-- 2. HASHING: bcrypt.GenerateFromPassword(password, cost=10) during user creation
-- 3. STORED HASH: Below hash value (one-way, cannot be reversed/unhashed)
-- 4. VERIFICATION: bcrypt.CompareHashAndPassword(storedHash, plaintext) during login
--
-- Hash Generation (for transparency):
--   Original Password: admin123
--   Algorithm: bcrypt with cost=10
--   Generated Hash: $2a$10$59d1WupSPuEYynVUaZOwhOOeOVg5.TsCcNK8oN.OaKzzbg.M5KvC.
--   (Verified: bcrypt.CompareHashAndPassword returns nil for this password)
-- ============================================================================

INSERT INTO users (username, password_hash, role, is_active)
VALUES ('admin', '$2a$10$59d1WupSPuEYynVUaZOwhOOeOVg5.TsCcNK8oN.OaKzzbg.M5KvC.', 'admin', true)
ON CONFLICT (username) DO NOTHING;