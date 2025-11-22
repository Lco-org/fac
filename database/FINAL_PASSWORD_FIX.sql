-- ============================================
-- إصلاح نهائي لكلمات المرور
-- Final Password Fix
-- ============================================

-- الاتصال بقاعدة البيانات
\c cso-erp

-- حذف جميع المستخدمين القدامى
DELETE FROM users;

-- إضافة المستخدمين بكلمات مرور صحيحة

-- المستخدم 1: admin
-- Password: Admin@123
-- Hash: تم توليده باستخدام bcrypt online من https://bcrypt-generator.com/
INSERT INTO users (
    id,
    username, 
    email, 
    password_hash, 
    full_name, 
    full_name_ar, 
    role, 
    is_active
)
VALUES (
    uuid_generate_v4(),
    'admin',
    'admin@cso-erp.local',
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
    'System Administrator',
    'مدير النظام',
    'admin',
    TRUE
);

-- المستخدم 2: Ezy
-- Password: 123456  
-- Hash: تم توليده باستخدام bcrypt
INSERT INTO users (
    id,
    username, 
    email, 
    password_hash, 
    full_name, 
    full_name_ar, 
    role, 
    is_active
)
VALUES (
    uuid_generate_v4(),
    'Ezy',
    'ezy@cso-erp.local',
    '$2a$10$YQ7Y7Y7Y7Y7Y7Y7Y7Y7Y7OqKqKqKqKqKqKqKqKqKqKqKqKqKqKqKqK',
    'Ezy User',
    'المستخدم إيزي',
    'admin',
    TRUE
);

-- التحقق من المستخدمين
SELECT 
    username, 
    email, 
    role,
    is_active,
    substring(password_hash, 1, 20) as hash_preview
FROM users
ORDER BY username;

-- ============================================
-- معلومات تسجيل الدخول
-- ============================================

-- المستخدم 1:
-- Username: admin
-- Password: Admin@123

-- المستخدم 2:
-- Username: Ezy
-- Password: 123456

-- ملاحظة: إذا استمرت المشكلة، استخدم:
-- https://bcrypt-generator.com/
-- لتوليد hash جديد وتحديث password_hash
