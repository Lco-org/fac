-- =====================================================
-- ERP Master Database Schema
-- Purpose: Multi-Tenant SaaS Platform Management
-- For: Super System Admin (SSA) Only
-- =====================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================
-- 1. ORGANIZATIONS TABLE (المنظمات/المستأجرين)
-- =====================================================
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Basic Info
    name VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255),
    slug VARCHAR(100) UNIQUE NOT NULL, -- URL-friendly identifier
    
    -- Domain & Database
    primary_domain VARCHAR(255) UNIQUE NOT NULL,
    database_name VARCHAR(100) UNIQUE NOT NULL,
    
    -- Contact
    contact_email VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50),
    
    -- Status: pending_verification, active, suspended, deleted
    status VARCHAR(50) DEFAULT 'pending_verification' NOT NULL,
    
    -- Subscription (toggleable feature)
    subscription_plan_id UUID,
    subscription_status VARCHAR(50) DEFAULT 'trial', -- trial, active, expired, cancelled
    subscription_expires_at TIMESTAMP,
    
    -- Metadata
    logo_url VARCHAR(500),
    settings JSONB DEFAULT '{}',
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    suspended_at TIMESTAMP,
    
    -- Created by Super System Admin
    created_by UUID NOT NULL
);

-- Indexes for organizations
CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations(status);
CREATE INDEX IF NOT EXISTS idx_organizations_domain ON organizations(primary_domain);
CREATE INDEX IF NOT EXISTS idx_organizations_slug ON organizations(slug);

-- =====================================================
-- 2. ALLOWED DOMAINS TABLE (الدومينات المسموح بها)
-- =====================================================
CREATE TABLE IF NOT EXISTS allowed_domains (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    domain VARCHAR(255) UNIQUE NOT NULL,
    is_primary BOOLEAN DEFAULT false NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    
    -- Verification
    verified BOOLEAN DEFAULT false NOT NULL,
    verified_at TIMESTAMP,
    verification_method VARCHAR(50), -- email, dns_txt
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Indexes for allowed_domains
CREATE INDEX IF NOT EXISTS idx_allowed_domains_org ON allowed_domains(organization_id);
CREATE INDEX IF NOT EXISTS idx_allowed_domains_domain ON allowed_domains(domain);
CREATE INDEX IF NOT EXISTS idx_allowed_domains_active ON allowed_domains(is_active);

-- =====================================================
-- 3. DOMAIN VERIFICATIONS TABLE (التحقق من الدومين)
-- =====================================================
CREATE TABLE IF NOT EXISTS domain_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    
    email VARCHAR(255) NOT NULL,
    hash_code VARCHAR(64) NOT NULL,
    
    -- Expiry: 5 minutes
    expires_at TIMESTAMP NOT NULL,
    
    -- Status
    verified BOOLEAN DEFAULT false NOT NULL,
    verified_at TIMESTAMP,
    attempts INT DEFAULT 0 NOT NULL,
    max_attempts INT DEFAULT 3 NOT NULL,
    
    -- IP tracking
    request_ip INET,
    verify_ip INET,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Indexes for domain_verifications
CREATE INDEX IF NOT EXISTS idx_domain_verifications_org ON domain_verifications(organization_id);
CREATE INDEX IF NOT EXISTS idx_domain_verifications_expires ON domain_verifications(expires_at);
CREATE INDEX IF NOT EXISTS idx_domain_verifications_code ON domain_verifications(hash_code);

-- =====================================================
-- 4. SUPER ADMIN AUDIT LOG (سجل عمليات المدير الأعلى)
-- =====================================================
CREATE TABLE IF NOT EXISTS super_admin_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    -- Actor (Super System Admin)
    admin_id UUID NOT NULL,
    admin_email VARCHAR(255) NOT NULL,
    
    -- Action details
    action VARCHAR(100) NOT NULL, -- create_org, suspend_org, activate_org, etc.
    resource_type VARCHAR(50) NOT NULL, -- organization, domain, user
    resource_id UUID,
    
    -- Change tracking
    old_values JSONB,
    new_values JSONB,
    
    -- Request metadata
    ip_address INET NOT NULL,
    user_agent TEXT,
    request_id VARCHAR(100),
    
    -- Blockchain-style integrity
    previous_hash VARCHAR(64),
    current_hash VARCHAR(64),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Indexes for super_admin_audit
CREATE INDEX IF NOT EXISTS idx_super_admin_audit_admin ON super_admin_audit(admin_id);
CREATE INDEX IF NOT EXISTS idx_super_admin_audit_action ON super_admin_audit(action);
CREATE INDEX IF NOT EXISTS idx_super_admin_audit_resource ON super_admin_audit(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_super_admin_audit_created ON super_admin_audit(created_at);

-- =====================================================
-- 5. SUBSCRIPTION PLANS TABLE (خطط الاشتراك - قابلة للإيقاف)
-- =====================================================
CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    name VARCHAR(100) NOT NULL,
    name_ar VARCHAR(100),
    code VARCHAR(50) UNIQUE NOT NULL, -- basic, pro, enterprise
    
    -- Pricing (0 = free)
    price_monthly DECIMAL(10, 2) DEFAULT 0,
    price_yearly DECIMAL(10, 2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'LYD',
    
    -- Features & Limits
    max_users INT DEFAULT 5,
    max_storage_gb INT DEFAULT 1,
    features JSONB DEFAULT '[]', -- list of feature codes
    
    -- Status
    is_active BOOLEAN DEFAULT true NOT NULL,
    is_default BOOLEAN DEFAULT false NOT NULL,
    
    -- Display
    sort_order INT DEFAULT 0,
    description TEXT,
    description_ar TEXT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Insert default plans
INSERT INTO subscription_plans (code, name, name_ar, max_users, max_storage_gb, is_default, sort_order) VALUES
    ('free', 'Free', 'مجاني', 3, 1, true, 1),
    ('basic', 'Basic', 'أساسي', 10, 5, false, 2),
    ('pro', 'Professional', 'احترافي', 50, 25, false, 3),
    ('enterprise', 'Enterprise', 'مؤسسي', -1, -1, false, 4) -- -1 = unlimited
ON CONFLICT (code) DO NOTHING;

-- =====================================================
-- 6. SYSTEM SETTINGS TABLE (إعدادات النظام الرئيسي)
-- =====================================================
CREATE TABLE IF NOT EXISTS system_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT,
    value_type VARCHAR(20) DEFAULT 'string', -- string, boolean, number, json
    
    category VARCHAR(50) DEFAULT 'general',
    description TEXT,
    
    is_public BOOLEAN DEFAULT false, -- Can tenants see this?
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Insert default settings
INSERT INTO system_settings (key, value, value_type, category, description) VALUES
    ('billing_enabled', 'false', 'boolean', 'billing', 'Enable/disable billing system'),
    ('max_organizations', '100', 'number', 'limits', 'Maximum number of organizations'),
    ('verification_code_expiry_minutes', '5', 'number', 'security', 'Verification code expiry in minutes'),
    ('max_verification_attempts', '3', 'number', 'security', 'Max verification code attempts'),
    ('allowed_super_admin_ips', '["127.0.0.1", "::1"]', 'json', 'security', 'IPs allowed for Super Admin access')
ON CONFLICT (key) DO NOTHING;

-- =====================================================
-- 7. SUPER SYSTEM ADMINS TABLE (مديرو النظام الأعلى)
-- =====================================================
CREATE TABLE IF NOT EXISTS super_system_admins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    
    full_name VARCHAR(255) NOT NULL,
    
    -- 2FA (mandatory for SSA)
    two_factor_enabled BOOLEAN DEFAULT true NOT NULL,
    two_factor_secret VARCHAR(255),
    
    -- Security
    token_version INT DEFAULT 0 NOT NULL,
    last_login TIMESTAMP,
    last_login_ip INET,
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMP,
    
    -- Status
    is_active BOOLEAN DEFAULT true NOT NULL,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- =====================================================
-- HELPER FUNCTIONS
-- =====================================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to tables with updated_at
CREATE TRIGGER update_organizations_updated_at
    BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscription_plans_updated_at
    BEFORE UPDATE ON subscription_plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_settings_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_super_system_admins_updated_at
    BEFORE UPDATE ON super_system_admins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- COMMENTS
-- =====================================================
COMMENT ON TABLE organizations IS 'Multi-tenant organizations managed by Super System Admin';
COMMENT ON TABLE allowed_domains IS 'Domains allowed for each organization login';
COMMENT ON TABLE domain_verifications IS 'Email verification codes for domain ownership';
COMMENT ON TABLE super_admin_audit IS 'Tamper-proof audit log for SSA actions';
COMMENT ON TABLE subscription_plans IS 'Subscription plans (feature can be disabled)';
COMMENT ON TABLE system_settings IS 'Global system settings for SSA';
COMMENT ON TABLE super_system_admins IS 'Super System Admin users with B8FB17 access';
