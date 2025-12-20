-- =====================================================
-- Organization Deletion Requests Table
-- For secure organization deletion with OTP verification
-- =====================================================

CREATE TABLE IF NOT EXISTS org_deletion_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    admin_id UUID NOT NULL,
    code_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    attempts INT DEFAULT 0 NOT NULL,
    max_attempts INT DEFAULT 3 NOT NULL,
    request_ip INET,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_org_deletion_requests_org ON org_deletion_requests(organization_id);
CREATE INDEX IF NOT EXISTS idx_org_deletion_requests_expires ON org_deletion_requests(expires_at);

-- Comment
COMMENT ON TABLE org_deletion_requests IS 'Temporary storage for organization deletion OTP verification';
