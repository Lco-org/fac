-- Add subscription_start_date to organizations
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS subscription_start_date TIMESTAMP;

-- Insert default subscription plans if not exist
INSERT INTO subscription_plans (name, code, price_monthly, max_users, max_storage_gb, features, is_default)
VALUES 
    ('Basic', 'basic', 0, 5, 1, '["invoices", "reports"]'::jsonb, true),
    ('Pro', 'pro', 50, 20, 10, '["invoices", "reports", "hr", "inventory"]'::jsonb, false),
    ('Enterprise', 'enterprise', 200, 100, 100, '["*"]'::jsonb, false)
ON CONFLICT (code) DO NOTHING;
