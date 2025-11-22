-- ============================================
-- CSO Financial ERP System - Migration 002
-- Add Employees and Members Tables
-- ============================================

-- Employees table
CREATE TABLE IF NOT EXISTS employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_number VARCHAR(50) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    full_name_ar VARCHAR(255),
    position VARCHAR(255),
    position_ar VARCHAR(255),
    department VARCHAR(255),
    department_ar VARCHAR(255),
    phone VARCHAR(50),
    email VARCHAR(255),
    salary NUMERIC(15, 2),
    join_date DATE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id)
);

CREATE INDEX idx_employees_number ON employees(employee_number);
CREATE INDEX idx_employees_active ON employees(is_active);
CREATE INDEX idx_employees_department ON employees(department);

-- Members table
CREATE TABLE IF NOT EXISTS members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    membership_number VARCHAR(50) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    full_name_ar VARCHAR(255),
    member_type VARCHAR(50) DEFAULT 'regular',
    phone VARCHAR(50),
    email VARCHAR(255),
    address TEXT,
    address_ar TEXT,
    join_date DATE,
    status VARCHAR(50) DEFAULT 'active',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by UUID REFERENCES users(id)
);

CREATE INDEX idx_members_number ON members(membership_number);
CREATE INDEX idx_members_status ON members(status);
CREATE INDEX idx_members_type ON members(member_type);
CREATE INDEX idx_members_active ON members(is_active);

-- Update triggers
CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_members_updated_at BEFORE UPDATE ON members
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE employees IS 'Organization employees with salary and position information';
COMMENT ON TABLE members IS 'Affiliated members with membership tracking';
