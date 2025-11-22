-- ============================================
-- CSO Financial ERP System - Database Schema
-- PostgreSQL Migration Script v1.0 - CORRECTED
-- ============================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- CUSTOM TYPES
-- ============================================

CREATE TYPE user_role AS ENUM ('admin', 'finance', 'auditor');
CREATE TYPE container_type AS ENUM ('cash', 'bank');
CREATE TYPE transaction_type AS ENUM (
    'income_grant',
    'income_donation',
    'disbursement_custody',
    'disbursement_salary',
    'disbursement_operational',
    'disbursement_donation',
    'disbursement_bonus',
    'procurement',
    'transfer'
);
CREATE TYPE transaction_status AS ENUM ('pending', 'completed', 'reversed');
CREATE TYPE currency_code AS ENUM ('LYD', 'USD', 'EUR', 'GBP');

-- ============================================
-- TABLES
-- ============================================

-- Users table with secure authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    full_name_ar VARCHAR(255),
    role user_role NOT NULL DEFAULT 'finance',
    is_active BOOLEAN DEFAULT TRUE,
    last_login TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT username_min_length CHECK (char_length(username) >= 3),
    CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- Projects (Cost Centers)
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255),
    description TEXT,
    donor_name VARCHAR(255),
    donor_name_ar VARCHAR(255),
    start_date DATE,
    end_date DATE,
    budget NUMERIC(15, 2),
    currency currency_code DEFAULT 'LYD',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id)
);

CREATE INDEX idx_projects_code ON projects(code);
CREATE INDEX idx_projects_active ON projects(is_active);

-- Financial Containers (Cash Boxes and Bank Accounts)
CREATE TABLE financial_containers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255),
    type container_type NOT NULL,
    currency currency_code DEFAULT 'LYD',
    balance NUMERIC(15, 2) DEFAULT 0.00,
    
    -- Bank-specific fields
    bank_name VARCHAR(255),
    bank_name_ar VARCHAR(255),
    account_number VARCHAR(100),
    iban VARCHAR(34),
    swift_code VARCHAR(11),
    
    -- Cash-specific fields
    location VARCHAR(255),
    location_ar VARCHAR(255),
    
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id),
    
    CONSTRAINT bank_fields_required CHECK (
        (type = 'bank' AND bank_name IS NOT NULL AND account_number IS NOT NULL) OR
        (type = 'cash')
    )
);

CREATE INDEX idx_containers_type ON financial_containers(type);
CREATE INDEX idx_containers_active ON financial_containers(is_active);
CREATE INDEX idx_containers_currency ON financial_containers(currency);

-- Transactions (All financial movements)
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_number VARCHAR(50) UNIQUE NOT NULL,
    type transaction_type NOT NULL,
    status transaction_status DEFAULT 'completed',
    
    -- Financial details
    amount NUMERIC(15, 2) NOT NULL CHECK (amount > 0),
    currency currency_code DEFAULT 'LYD',
    exchange_rate NUMERIC(10, 4) DEFAULT 1.0000,
    amount_local NUMERIC(15, 2),
    
    -- Container and project (REQUIRED)
    container_id UUID NOT NULL REFERENCES financial_containers(id),
    project_id UUID NOT NULL REFERENCES projects(id),
    
    -- Transaction details
    description TEXT NOT NULL,
    description_ar TEXT,
    payee_payer VARCHAR(255),
    payee_payer_ar VARCHAR(255),
    reference_number VARCHAR(100),
    
    -- Deficit handling
    is_deficit BOOLEAN DEFAULT FALSE,
    deficit_approved_by UUID REFERENCES users(id),
    deficit_warning_shown BOOLEAN DEFAULT FALSE,
    balance_before NUMERIC(15, 2),
    balance_after NUMERIC(15, 2),
    
    -- Attachments
    attachment_url TEXT,
    
    -- Metadata
    transaction_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID NOT NULL REFERENCES users(id),
    
    -- Reversal tracking
    reversed_at TIMESTAMP WITH TIME ZONE,
    reversed_by UUID REFERENCES users(id),
    reversal_reason TEXT,
    original_transaction_id UUID REFERENCES transactions(id)
);

CREATE INDEX idx_transactions_number ON transactions(transaction_number);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_container ON transactions(container_id);
CREATE INDEX idx_transactions_project ON transactions(project_id);
CREATE INDEX idx_transactions_date ON transactions(transaction_date);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_created_by ON transactions(created_by);

-- Audit Log
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    changed_by UUID REFERENCES users(id),
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT
);

CREATE INDEX idx_audit_table_record ON audit_log(table_name, record_id);
CREATE INDEX idx_audit_changed_at ON audit_log(changed_at);
CREATE INDEX idx_audit_changed_by ON audit_log(changed_by);

-- ============================================
-- FUNCTIONS AND TRIGGERS
-- ============================================

-- Update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_containers_updated_at BEFORE UPDATE ON financial_containers
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_transactions_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Generate transaction number
CREATE OR REPLACE FUNCTION generate_transaction_number()
RETURNS TRIGGER AS $$
DECLARE
    next_num INTEGER;
    year_part VARCHAR(4);
BEGIN
    year_part := TO_CHAR(CURRENT_DATE, 'YYYY');
    
    SELECT COALESCE(MAX(CAST(SUBSTRING(transaction_number FROM 9) AS INTEGER)), 0) + 1
    INTO next_num
    FROM transactions
    WHERE transaction_number LIKE 'TX-' || year_part || '-%';
    
    NEW.transaction_number := 'TX-' || year_part || '-' || LPAD(next_num::TEXT, 5, '0');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER generate_transaction_number_trigger
    BEFORE INSERT ON transactions
    FOR EACH ROW
    WHEN (NEW.transaction_number IS NULL OR NEW.transaction_number = '')
    EXECUTE FUNCTION generate_transaction_number();

-- Calculate local amount
CREATE OR REPLACE FUNCTION calculate_local_amount()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.amount_local IS NULL THEN
        NEW.amount_local := NEW.amount * NEW.exchange_rate;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER calculate_local_amount_trigger
    BEFORE INSERT OR UPDATE ON transactions
    FOR EACH ROW
    EXECUTE FUNCTION calculate_local_amount();

-- ============================================
-- VIEWS
-- ============================================

-- Daily Financial Position
CREATE OR REPLACE VIEW v_daily_position AS
SELECT 
    fc.id,
    fc.name,
    fc.name_ar,
    fc.type,
    fc.currency,
    fc.balance,
    CASE 
        WHEN fc.balance < 0 THEN 'deficit'
        WHEN fc.balance = 0 THEN 'zero'
        ELSE 'surplus'
    END AS balance_status,
    fc.bank_name,
    fc.bank_name_ar,
    fc.account_number,
    fc.location,
    fc.location_ar,
    fc.is_active,
    -- Today's movements
    COALESCE(SUM(CASE 
        WHEN t.type IN ('income_grant', 'income_donation') 
        AND t.transaction_date = CURRENT_DATE 
        THEN t.amount 
        ELSE 0 
    END), 0) AS today_income,
    COALESCE(SUM(CASE 
        WHEN t.type::TEXT LIKE 'disbursement_%' 
        AND t.transaction_date = CURRENT_DATE 
        THEN t.amount 
        ELSE 0 
    END), 0) AS today_expenses
FROM financial_containers fc
LEFT JOIN transactions t ON fc.id = t.container_id AND t.status = 'completed'
WHERE fc.is_active = TRUE
GROUP BY fc.id, fc.name, fc.name_ar, fc.type, fc.currency, fc.balance, 
         fc.bank_name, fc.bank_name_ar, fc.account_number, fc.location, fc.location_ar, fc.is_active
ORDER BY fc.type, fc.name;

-- Project Financial Summary
CREATE OR REPLACE VIEW v_project_summary AS
SELECT 
    p.id,
    p.code,
    p.name,
    p.name_ar,
    p.currency,
    p.donor_name,
    p.donor_name_ar,
    COALESCE(SUM(CASE 
        WHEN t.type IN ('income_grant', 'income_donation') 
        THEN t.amount_local 
        ELSE 0 
    END), 0) AS total_income,
    COALESCE(SUM(CASE 
        WHEN t.type::TEXT LIKE 'disbursement_%' 
        THEN t.amount_local 
        ELSE 0 
    END), 0) AS total_expenses,
    COALESCE(SUM(CASE 
        WHEN t.type IN ('income_grant', 'income_donation') 
        THEN t.amount_local 
        ELSE -t.amount_local 
    END), 0) AS net_position,
    COUNT(t.id) AS transaction_count
FROM projects p
LEFT JOIN transactions t ON p.id = t.project_id AND t.status = 'completed'
WHERE p.is_active = TRUE
GROUP BY p.id, p.code, p.name, p.name_ar, p.currency, p.donor_name, p.donor_name_ar;

-- Deficit Report
CREATE OR REPLACE VIEW v_deficit_report AS
SELECT 
    fc.id,
    fc.name,
    fc.name_ar,
    fc.type,
    fc.currency,
    fc.balance AS deficit_amount,
    COUNT(t.id) AS deficit_transactions_count,
    MIN(t.transaction_date) AS first_deficit_date,
    MAX(t.transaction_date) AS last_deficit_date
FROM financial_containers fc
LEFT JOIN transactions t ON fc.id = t.container_id AND t.is_deficit = TRUE AND t.status = 'completed'
WHERE fc.balance < 0 AND fc.is_active = TRUE
GROUP BY fc.id, fc.name, fc.name_ar, fc.type, fc.currency, fc.balance
ORDER BY fc.balance ASC;

-- ============================================
-- INITIAL DATA
-- ============================================

-- Insert admin user
-- Username: admin
-- Password: Admin@123
-- Hash generated with bcrypt cost 10
INSERT INTO users (username, email, password_hash, full_name, full_name_ar, role, is_active)
VALUES (
    'admin',
    'admin@cso-erp.local',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi',
    'System Administrator',
    'مدير النظام',
    'admin',
    TRUE
);

-- Insert Ezy user
-- Username: Ezy
-- Password: 123456
INSERT INTO users (username, email, password_hash, full_name, full_name_ar, role, is_active)
VALUES (
    'Ezy',
    'ezy@cso-erp.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'Ezy User',
    'المستخدم إيزي',
    'admin',
    TRUE
);

-- Insert General Administration project
INSERT INTO projects (code, name, name_ar, description, is_active, created_by)
VALUES (
    'ADMIN',
    'General Administration',
    'الإدارة العامة',
    'General administrative expenses not tied to specific projects',
    TRUE,
    (SELECT id FROM users WHERE username = 'admin')
);

-- ============================================
-- COMMENTS
-- ============================================

COMMENT ON TABLE users IS 'System users with role-based access control';
COMMENT ON TABLE projects IS 'Cost centers for tracking project-specific expenses';
COMMENT ON TABLE financial_containers IS 'Cash boxes and bank accounts with deficit support';
COMMENT ON TABLE transactions IS 'All financial transactions with complete audit trail';
COMMENT ON TABLE audit_log IS 'Complete audit trail for compliance and security';

COMMENT ON COLUMN financial_containers.balance IS 'Current balance - CAN BE NEGATIVE (deficit/overdraft)';
COMMENT ON COLUMN transactions.is_deficit IS 'TRUE if this transaction created a negative balance';
COMMENT ON COLUMN transactions.amount_local IS 'Amount converted to LYD using exchange_rate';
COMMENT ON COLUMN transactions.balance_before IS 'Container balance before this transaction';
COMMENT ON COLUMN transactions.balance_after IS 'Container balance after this transaction';

-- ============================================
-- LOGIN CREDENTIALS
-- ============================================

-- User 1:
-- Username: admin
-- Password: Admin@123
-- Role: admin

-- User 2:
-- Username: Ezy
-- Password: 123456
-- Role: admin