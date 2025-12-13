-- ============================================================================
-- TENANT DATABASE SCHEMA ADDITIONS
-- Run this on each tenant database (or add to the template cso-erp)
-- ============================================================================

-- Email Templates Table
-- Allows each organization to customize their email templates
CREATE TABLE IF NOT EXISTS email_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_key VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255),
    subject VARCHAR(500) NOT NULL,
    subject_ar VARCHAR(500),
    body_html TEXT NOT NULL,
    body_text TEXT,
    description TEXT,
    variables JSONB DEFAULT '[]'::jsonb,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

-- Email template keys that can be customized
COMMENT ON TABLE email_templates IS 'Customizable email templates for the organization';
COMMENT ON COLUMN email_templates.template_key IS 'Unique key: welcome_user, password_reset, 2fa_code, transaction_receipt, etc';
COMMENT ON COLUMN email_templates.variables IS 'JSON array of variable names used in template: ["user_name", "code", "amount"]';

-- Create indexes
CREATE INDEX idx_email_templates_key ON email_templates(template_key);
CREATE INDEX idx_email_templates_active ON email_templates(is_active) WHERE is_active = true;

-- SMTP Settings Table
-- Allows each organization to configure their own SMTP server
CREATE TABLE IF NOT EXISTS smtp_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL DEFAULT 'Default SMTP',
    smtp_host VARCHAR(255) NOT NULL,
    smtp_port INTEGER NOT NULL DEFAULT 587,
    username VARCHAR(255),
    password_encrypted TEXT,
    from_address VARCHAR(255) NOT NULL,
    from_name VARCHAR(255),
    use_tls BOOLEAN DEFAULT true,
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT true,
    test_email_sent_at TIMESTAMP WITH TIME ZONE,
    last_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by UUID REFERENCES users(id) ON DELETE SET NULL
);

COMMENT ON TABLE smtp_settings IS 'SMTP configuration for organization email sending';
COMMENT ON COLUMN smtp_settings.password_encrypted IS 'AES-256-GCM encrypted SMTP password';
COMMENT ON COLUMN smtp_settings.is_default IS 'Only one SMTP config can be default';

-- Ensure only one default SMTP config
CREATE UNIQUE INDEX idx_smtp_settings_default ON smtp_settings(is_active, is_default) WHERE is_default = true AND is_active = true;

-- Organization Settings Table (for tenant-specific settings)
CREATE TABLE IF NOT EXISTS organization_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    setting_key VARCHAR(100) NOT NULL UNIQUE,
    setting_value TEXT,
    setting_type VARCHAR(50) DEFAULT 'string',
    description TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

COMMENT ON TABLE organization_settings IS 'Key-value store for organization-specific settings';
COMMENT ON COLUMN organization_settings.setting_type IS 'Type: string, number, boolean, json';

-- Insert default email templates
INSERT INTO email_templates (template_key, name, name_ar, subject, subject_ar, body_html, variables, description) VALUES
(
    'welcome_user',
    'Welcome New User',
    'مرحباً بالمستخدم الجديد',
    'Welcome to {{org_name}}!',
    'مرحباً بك في {{org_name}}!',
    '<h1>مرحباً {{user_name}}</h1><p>تم إنشاء حسابك بنجاح.</p>',
    '["user_name", "org_name", "username", "login_url"]'::jsonb,
    'Sent when a new user is created'
),
(
    '2fa_code',
    '2FA Verification Code',
    'رمز التحقق بخطوتين',
    'Your verification code: {{code}}',
    'رمز التحقق الخاص بك: {{code}}',
    '<h1>رمز التحقق</h1><p>رمزك هو: <strong>{{code}}</strong></p><p>صالح لمدة 10 دقائق.</p>',
    '["code", "user_name", "expiry_minutes"]'::jsonb,
    'Two-factor authentication code'
),
(
    'password_reset',
    'Password Reset',
    'إعادة تعيين كلمة المرور',
    'Password Reset Request',
    'طلب إعادة تعيين كلمة المرور',
    '<h1>إعادة تعيين كلمة المرور</h1><p>اضغط على الرابط التالي: {{reset_link}}</p>',
    '["user_name", "reset_link", "expiry_hours"]'::jsonb,
    'Password reset request email'
),
(
    'transaction_receipt',
    'Transaction Receipt',
    'إيصال معاملة',
    'Transaction Receipt #{{transaction_id}}',
    'إيصال معاملة رقم {{transaction_id}}',
    '<h1>إيصال معاملة</h1><p>رقم المعاملة: {{transaction_id}}</p><p>المبلغ: {{amount}}</p>',
    '["transaction_id", "amount", "date", "description", "type"]'::jsonb,
    'Transaction receipt notification'
)
ON CONFLICT (template_key) DO NOTHING;

-- Set updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_email_templates_updated_at
    BEFORE UPDATE ON email_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_smtp_settings_updated_at
    BEFORE UPDATE ON smtp_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_organization_settings_updated_at
    BEFORE UPDATE ON organization_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
