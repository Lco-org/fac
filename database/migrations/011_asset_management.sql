-- =====================================================
-- Asset Management Migration
-- Version: 011
-- تطبق على: cso-erp (القالب) + جميع قواعد المستأجرين
-- =====================================================

-- 1. جدول سجل الـ migrations المطبقة
CREATE TABLE IF NOT EXISTS applied_migrations (
    id SERIAL PRIMARY KEY,
    migration_name VARCHAR(255) UNIQUE NOT NULL,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- التحقق من عدم تطبيق هذا الـ migration مسبقاً
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM applied_migrations WHERE migration_name = '011_asset_management') THEN
        RAISE NOTICE 'Migration 011_asset_management already applied, skipping...';
        RETURN;
    END IF;

    -- 2. فئات الأصول
    CREATE TABLE IF NOT EXISTS asset_categories (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        name VARCHAR(100) NOT NULL,
        name_ar VARCHAR(100) NOT NULL,
        icon VARCHAR(50) DEFAULT 'box',
        color VARCHAR(7) DEFAULT '#6366f1',
        sort_order INT DEFAULT 0,
        is_active BOOLEAN DEFAULT true,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    -- 3. أنواع الأصول
    CREATE TABLE IF NOT EXISTS asset_types (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        category_id UUID NOT NULL REFERENCES asset_categories(id) ON DELETE CASCADE,
        name VARCHAR(100) NOT NULL,
        name_ar VARCHAR(100) NOT NULL,
        unit VARCHAR(50) DEFAULT 'قطعة',
        depreciation_years INT,
        is_active BOOLEAN DEFAULT true,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    -- 4. الأصول
    CREATE TABLE IF NOT EXISTS assets (
        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
        asset_type_id UUID NOT NULL REFERENCES asset_types(id),
        project_id UUID REFERENCES projects(id),
        transaction_id UUID REFERENCES transactions(id),
        
        -- التعريف
        asset_code VARCHAR(50) UNIQUE,
        serial_number VARCHAR(100),
        barcode VARCHAR(100),
        
        -- الوصف
        name VARCHAR(255),
        name_ar VARCHAR(255),
        description TEXT,
        description_ar TEXT,
        
        -- القيمة
        quantity INT DEFAULT 1,
        unit_price DECIMAL(15,2) DEFAULT 0,
        total_value DECIMAL(15,2) DEFAULT 0,
        currency VARCHAR(3) DEFAULT 'LYD',
        
        -- الحالة
        condition VARCHAR(20) DEFAULT 'new',
        status VARCHAR(20) DEFAULT 'in_use',
        
        -- التتبع
        acquisition_date DATE DEFAULT CURRENT_DATE,
        acquisition_type VARCHAR(20) NOT NULL,
        supplier VARCHAR(255),
        warranty_expires DATE,
        
        -- الموقع
        location VARCHAR(255),
        assigned_to UUID REFERENCES employees(id),
        
        -- ملاحظات
        notes TEXT,
        
        -- النظام
        is_active BOOLEAN DEFAULT true,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        created_by UUID NOT NULL
    );

    -- الفهارس
    CREATE INDEX IF NOT EXISTS idx_assets_type ON assets(asset_type_id);
    CREATE INDEX IF NOT EXISTS idx_assets_project ON assets(project_id);
    CREATE INDEX IF NOT EXISTS idx_assets_transaction ON assets(transaction_id);
    CREATE INDEX IF NOT EXISTS idx_assets_code ON assets(asset_code);
    CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);
    CREATE INDEX IF NOT EXISTS idx_asset_types_category ON asset_types(category_id);

    -- 5. إدراج الفئات الافتراضية
    INSERT INTO asset_categories (name, name_ar, icon, color, sort_order) VALUES
        ('Real Estate', 'عقارات', 'building', '#10B981', 1),
        ('Office Furniture', 'أثاث مكتبي', 'armchair', '#F59E0B', 2),
        ('Electronics', 'أجهزة إلكترونية', 'laptop', '#3B82F6', 3),
        ('Servers & Network', 'سيرفرات واتصالات', 'server', '#8B5CF6', 4),
        ('Vehicles', 'مركبات', 'car', '#EF4444', 5),
        ('Office Supplies', 'مستلزمات مكتبية', 'paperclip', '#6B7280', 6),
        ('Other', 'أخرى', 'package', '#64748B', 99)
    ON CONFLICT DO NOTHING;

    -- 6. إدراج الأنواع الافتراضية
    -- عقارات
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Land', 'أرض', 'متر مربع' FROM asset_categories WHERE name = 'Real Estate'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Building', 'مبنى', 'قطعة' FROM asset_categories WHERE name = 'Real Estate'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Office Space', 'مكتب', 'قطعة' FROM asset_categories WHERE name = 'Real Estate'
    ON CONFLICT DO NOTHING;

    -- أثاث مكتبي
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Desk', 'مكتب', 'قطعة' FROM asset_categories WHERE name = 'Office Furniture'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Chair', 'كرسي', 'قطعة' FROM asset_categories WHERE name = 'Office Furniture'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Cabinet', 'خزانة', 'قطعة' FROM asset_categories WHERE name = 'Office Furniture'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Meeting Table', 'طاولة اجتماعات', 'قطعة' FROM asset_categories WHERE name = 'Office Furniture'
    ON CONFLICT DO NOTHING;

    -- إلكترونيات
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Desktop Computer', 'حاسوب مكتبي', 'قطعة' FROM asset_categories WHERE name = 'Electronics'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Laptop', 'لابتوب', 'قطعة' FROM asset_categories WHERE name = 'Electronics'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Printer', 'طابعة', 'قطعة' FROM asset_categories WHERE name = 'Electronics'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Monitor', 'شاشة', 'قطعة' FROM asset_categories WHERE name = 'Electronics'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Phone', 'هاتف', 'قطعة' FROM asset_categories WHERE name = 'Electronics'
    ON CONFLICT DO NOTHING;

    -- سيرفرات واتصالات
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Server', 'سيرفر', 'قطعة' FROM asset_categories WHERE name = 'Servers & Network'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Router', 'راوتر', 'قطعة' FROM asset_categories WHERE name = 'Servers & Network'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Switch', 'سويتش', 'قطعة' FROM asset_categories WHERE name = 'Servers & Network'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'UPS', 'UPS', 'قطعة' FROM asset_categories WHERE name = 'Servers & Network'
    ON CONFLICT DO NOTHING;

    -- مركبات
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Car', 'سيارة', 'قطعة' FROM asset_categories WHERE name = 'Vehicles'
    ON CONFLICT DO NOTHING;
    
    INSERT INTO asset_types (category_id, name, name_ar, unit)
    SELECT id, 'Motorcycle', 'دراجة نارية', 'قطعة' FROM asset_categories WHERE name = 'Vehicles'
    ON CONFLICT DO NOTHING;

    -- تسجيل الـ migration
    INSERT INTO applied_migrations (migration_name) VALUES ('011_asset_management');
    
    RAISE NOTICE 'Migration 011_asset_management applied successfully!';
END $$;
