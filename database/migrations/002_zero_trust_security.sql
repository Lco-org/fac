-- =====================================================
-- Zero-Trust Security Migration
-- CSO ERP Financial System
-- =====================================================
-- This migration adds security enhancements to the existing database:
-- 1. Token versioning for Kill Switch (session invalidation)
-- 2. Two-Factor Authentication (2FA) via email
-- 3. Tamper-proof audit logs with blockchain-style chaining
-- =====================================================

BEGIN;

-- =====================================================
-- 1. USER TABLE ENHANCEMENTS
-- =====================================================

-- Add token_version for Kill Switch functionality
-- Incrementing this value invalidates all existing JWT tokens
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS token_version INTEGER DEFAULT 0 NOT NULL;

-- Add 2FA enabled flag (default: true for all users)
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN DEFAULT true NOT NULL;

-- Create index for faster token version lookups during authentication
CREATE INDEX IF NOT EXISTS idx_users_token_version 
    ON users(id, token_version);

-- =====================================================
-- 2. OTP CODES TABLE (for 2FA)
-- =====================================================

CREATE TABLE IF NOT EXISTS otp_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    attempts INTEGER DEFAULT 0 NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for OTP lookups
CREATE INDEX IF NOT EXISTS idx_otp_codes_user_id 
    ON otp_codes(user_id);
CREATE INDEX IF NOT EXISTS idx_otp_codes_expires_at 
    ON otp_codes(expires_at);

-- Function to cleanup expired OTPs
CREATE OR REPLACE FUNCTION cleanup_expired_otps()
RETURNS void AS $$
BEGIN
    DELETE FROM otp_codes WHERE expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- =====================================================
-- 3. TAMPER-PROOF AUDIT LOGS
-- =====================================================

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence BIGINT NOT NULL UNIQUE,
    action VARCHAR(50) NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    username VARCHAR(255),
    resource_type VARCHAR(100) NOT NULL,
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    previous_hash VARCHAR(64) NOT NULL,
    current_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
);

-- Indexes for audit log queries
CREATE INDEX IF NOT EXISTS idx_audit_logs_sequence 
    ON audit_logs(sequence);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id 
    ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action 
    ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at 
    ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource 
    ON audit_logs(resource_type, resource_id);

-- =====================================================
-- 4. AUDIT LOG IMMUTABILITY ENFORCEMENT
-- =====================================================

-- Function to prevent modifications to audit logs
CREATE OR REPLACE FUNCTION prevent_audit_modification()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit logs are immutable and cannot be modified or deleted.';
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to prevent UPDATE on audit_logs
DROP TRIGGER IF EXISTS prevent_audit_update ON audit_logs;
CREATE TRIGGER prevent_audit_update
    BEFORE UPDATE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_modification();

-- Trigger to prevent DELETE on audit_logs
DROP TRIGGER IF EXISTS prevent_audit_delete ON audit_logs;
CREATE TRIGGER prevent_audit_delete
    BEFORE DELETE ON audit_logs
    FOR EACH ROW
    EXECUTE FUNCTION prevent_audit_modification();

-- =====================================================
-- 5. DOCUMENTATION COMMENTS
-- =====================================================

COMMENT ON COLUMN users.token_version IS 
    'Incremented to invalidate all sessions (Kill Switch feature)';

COMMENT ON COLUMN users.two_factor_enabled IS 
    'Whether email 2FA is enabled for this user';

COMMENT ON TABLE otp_codes IS 
    'Temporary storage for email-based 2FA verification codes';

COMMENT ON COLUMN otp_codes.code_hash IS 
    'Argon2id hashed OTP code for secure comparison';

COMMENT ON COLUMN otp_codes.attempts IS 
    'Number of failed verification attempts (max 5)';

COMMENT ON TABLE audit_logs IS 
    'Tamper-proof audit log with blockchain-style chained hashing';

COMMENT ON COLUMN audit_logs.sequence IS 
    'Contiguous sequence number - gaps indicate deleted entries';

COMMENT ON COLUMN audit_logs.previous_hash IS 
    'SHA-256 hash of previous log entry (GENESIS for first)';

COMMENT ON COLUMN audit_logs.current_hash IS 
    'SHA-256 hash of this entry - modification breaks the chain';

COMMIT;

-- =====================================================
-- VERIFICATION QUERIES (Optional - for testing)
-- =====================================================
-- 
-- Check users table columns:
-- SELECT column_name, data_type FROM information_schema.columns 
-- WHERE table_name = 'users' AND column_name IN ('token_version', 'two_factor_enabled');
--
-- Check tables created:
-- SELECT table_name FROM information_schema.tables 
-- WHERE table_schema = 'public' AND table_name IN ('otp_codes', 'audit_logs');
--
-- Check triggers:
-- SELECT trigger_name FROM information_schema.triggers 
-- WHERE trigger_schema = 'public' AND event_object_table = 'audit_logs';
