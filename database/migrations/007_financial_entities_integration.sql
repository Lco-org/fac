-- Migration: 007_financial_entities_integration.sql
-- Description: Integrate employees and members as active financial entities
-- Features:
--   1. Link transactions to employees and members
--   2. Add payroll fields to employees (base_salary, monthly_budget)
--   3. Add membership validity fields to members
--   4. Create salary_deductions table
--   5. Create member_subscriptions table
--   6. Create view for membership status

BEGIN;

-- ============================================================================
-- 1. ADD EMPLOYEE AND MEMBER REFERENCES TO TRANSACTIONS
-- ============================================================================

-- Add employee_id column for salary, custody, and operational expense tracking
ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS employee_id UUID REFERENCES employees(id) ON DELETE SET NULL;

-- Add member_id column for member deposits and contributions
ALTER TABLE transactions
ADD COLUMN IF NOT EXISTS member_id UUID REFERENCES members(id) ON DELETE SET NULL;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_transactions_employee_id 
ON transactions(employee_id) WHERE employee_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_member_id 
ON transactions(member_id) WHERE member_id IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN transactions.employee_id IS 'Employee linked to this transaction (mandatory for salary/custody, optional for operational)';
COMMENT ON COLUMN transactions.member_id IS 'Member linked to this transaction (for member deposits/contributions)';

-- ============================================================================
-- 2. ADD PAYROLL FIELDS TO EMPLOYEES
-- ============================================================================

-- Add base salary field
ALTER TABLE employees
ADD COLUMN IF NOT EXISTS base_salary DECIMAL(15,2) DEFAULT 0.00;

-- Add monthly budget field
ALTER TABLE employees
ADD COLUMN IF NOT EXISTS monthly_budget DECIMAL(15,2) DEFAULT 0.00;

-- Add salary currency field
ALTER TABLE employees
ADD COLUMN IF NOT EXISTS salary_currency VARCHAR(3) DEFAULT 'LYD';

-- Add constraints (drop first if exists to avoid errors on re-run)
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_employees_base_salary'
    ) THEN
        ALTER TABLE employees
        ADD CONSTRAINT chk_employees_base_salary CHECK (base_salary >= 0);
    END IF;
    
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_employees_monthly_budget'
    ) THEN
        ALTER TABLE employees
        ADD CONSTRAINT chk_employees_monthly_budget CHECK (monthly_budget >= 0);
    END IF;
END $$;

-- Add comments
COMMENT ON COLUMN employees.base_salary IS 'Employee base salary amount';
COMMENT ON COLUMN employees.monthly_budget IS 'Monthly budget allocation for the employee';
COMMENT ON COLUMN employees.salary_currency IS 'Currency code for salary (LYD, USD, EUR, GBP)';

-- ============================================================================
-- 3. ADD MEMBERSHIP VALIDITY FIELDS TO MEMBERS
-- ============================================================================

-- Add valid_until field for membership expiry
ALTER TABLE members
ADD COLUMN IF NOT EXISTS valid_until TIMESTAMP WITH TIME ZONE;

-- Add last_renewal_date field
ALTER TABLE members
ADD COLUMN IF NOT EXISTS last_renewal_date TIMESTAMP WITH TIME ZONE;

-- Add subscription amount field
ALTER TABLE members
ADD COLUMN IF NOT EXISTS subscription_amount DECIMAL(15,2) DEFAULT 0.00;

-- Add subscription currency field
ALTER TABLE members
ADD COLUMN IF NOT EXISTS subscription_currency VARCHAR(3) DEFAULT 'LYD';

-- Add comments
COMMENT ON COLUMN members.valid_until IS 'Date when membership expires (1 year from last deposit/activation)';
COMMENT ON COLUMN members.last_renewal_date IS 'Date of last subscription renewal/payment';
COMMENT ON COLUMN members.subscription_amount IS 'Annual subscription amount';
COMMENT ON COLUMN members.subscription_currency IS 'Currency for subscription payments';

-- ============================================================================
-- 4. CREATE SALARY DEDUCTIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS salary_deductions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    month_year DATE NOT NULL, -- First day of the month (e.g., 2025-01-01)
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'LYD',
    reason TEXT NOT NULL,
    reason_ar TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    CONSTRAINT chk_salary_deductions_amount CHECK (amount > 0)
);

-- Create indexes for salary deductions
CREATE INDEX IF NOT EXISTS idx_salary_deductions_employee_id 
ON salary_deductions(employee_id);

CREATE INDEX IF NOT EXISTS idx_salary_deductions_month_year 
ON salary_deductions(month_year);

CREATE INDEX IF NOT EXISTS idx_salary_deductions_employee_month 
ON salary_deductions(employee_id, month_year);

COMMENT ON TABLE salary_deductions IS 'Tracks monthly salary deductions for employees with reasons';

-- ============================================================================
-- 5. CREATE MEMBER SUBSCRIPTIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS member_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    member_id UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    transaction_id UUID REFERENCES transactions(id) ON DELETE SET NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'LYD',
    payment_date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_until TIMESTAMP WITH TIME ZONE NOT NULL,
    notes TEXT,
    notes_ar TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    
    CONSTRAINT chk_member_subscriptions_amount CHECK (amount > 0),
    CONSTRAINT chk_member_subscriptions_validity CHECK (valid_until > valid_from)
);

-- Create indexes for member subscriptions
CREATE INDEX IF NOT EXISTS idx_member_subscriptions_member_id 
ON member_subscriptions(member_id);

CREATE INDEX IF NOT EXISTS idx_member_subscriptions_valid_until 
ON member_subscriptions(valid_until);

CREATE INDEX IF NOT EXISTS idx_member_subscriptions_transaction_id 
ON member_subscriptions(transaction_id) WHERE transaction_id IS NOT NULL;

COMMENT ON TABLE member_subscriptions IS 'Tracks member subscription payments with 365-day validity periods';

-- ============================================================================
-- 6. CREATE VIEW FOR MEMBERSHIP STATUS
-- ============================================================================

-- Drop view if exists to recreate
DROP VIEW IF EXISTS vw_members_with_status;

-- Create view with computed membership status (auto-freeze logic)
CREATE VIEW vw_members_with_status AS
SELECT 
    m.id,
    m.membership_number,
    m.full_name,
    m.full_name_ar,
    m.member_type,
    m.phone,
    m.email,
    m.address,
    m.address_ar,
    m.join_date,
    m.status as original_status,
    m.valid_until,
    m.last_renewal_date,
    m.subscription_amount,
    m.subscription_currency,
    m.is_active,
    m.created_at,
    m.updated_at,
    m.created_by,
    -- Computed status with auto-freeze logic
    CASE 
        WHEN m.is_active = FALSE THEN 'inactive'
        WHEN m.valid_until IS NULL THEN 'pending'
        WHEN m.valid_until < CURRENT_TIMESTAMP THEN 'frozen'
        WHEN m.valid_until >= CURRENT_TIMESTAMP THEN 'active'
        ELSE COALESCE(m.status, 'pending')
    END as computed_status,
    -- Flag for expiring soon (within 30 days)
    CASE 
        WHEN m.valid_until IS NOT NULL 
             AND m.valid_until >= CURRENT_TIMESTAMP 
             AND m.valid_until < CURRENT_TIMESTAMP + INTERVAL '30 days' 
        THEN TRUE
        ELSE FALSE
    END as expiring_soon,
    -- Days remaining until expiry
    CASE 
        WHEN m.valid_until IS NOT NULL AND m.valid_until >= CURRENT_TIMESTAMP
        THEN EXTRACT(DAY FROM (m.valid_until - CURRENT_TIMESTAMP))::INTEGER
        ELSE 0
    END as days_remaining
FROM members m;

COMMENT ON VIEW vw_members_with_status IS 'View showing members with computed membership status based on validity dates';

-- ============================================================================
-- 7. INSERT DASHBOARD WIDGET FOR EXPIRING MEMBERSHIPS
-- ============================================================================

INSERT INTO dashboard_widgets (widget_key, widget_name, widget_name_en, is_enabled, allowed_roles, display_order)
VALUES (
    'expiring_memberships',
    'الأعضاء المنتهية صلاحيتهم قريباً',
    'Expiring Memberships',
    TRUE,
    ARRAY['admin', 'finance'],
    8
)
ON CONFLICT (widget_key) DO NOTHING;

COMMIT;

-- ============================================================================
-- VERIFICATION QUERIES (Run after migration to verify)
-- ============================================================================

-- Check transactions table structure
-- SELECT column_name, data_type, is_nullable 
-- FROM information_schema.columns 
-- WHERE table_name = 'transactions' 
-- AND column_name IN ('employee_id', 'member_id');

-- Check employees table structure
-- SELECT column_name, data_type, is_nullable 
-- FROM information_schema.columns 
-- WHERE table_name = 'employees' 
-- AND column_name IN ('base_salary', 'monthly_budget', 'salary_currency');

-- Check members table structure
-- SELECT column_name, data_type, is_nullable 
-- FROM information_schema.columns 
-- WHERE table_name = 'members' 
-- AND column_name IN ('valid_until', 'last_renewal_date', 'subscription_amount');

-- Check new tables exist
-- SELECT table_name FROM information_schema.tables 
-- WHERE table_name IN ('salary_deductions', 'member_subscriptions');

-- Check view exists
-- SELECT table_name FROM information_schema.views 
-- WHERE table_name = 'vw_members_with_status';
