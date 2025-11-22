-- ============================================
-- إصلاح كلمات المرور - Fix User Passwords
-- ============================================

-- تحديث كلمة مرور المستخدم admin
-- Username: admin
-- Password: Admin@123
UPDATE users 
SET password_hash = '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'
WHERE username = 'admin';

-- إضافة أو تحديث مستخدم Ezy
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
)
ON CONFLICT (username) 
DO UPDATE SET password_hash = EXCLUDED.password_hash;

-- التحقق من المستخدمين
SELECT 
    username, 
    email, 
    role, 
    is_active,
    created_at
FROM users
ORDER BY created_at;

-- ============================================
-- معلومات تسجيل الدخول
-- ============================================

-- المستخدم 1:
-- Username: admin
-- Password: Admin@123
-- Role: admin

-- المستخدم 2:
-- Username: Ezy
-- Password: 123456
-- Role: admin
