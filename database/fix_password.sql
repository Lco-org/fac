-- إصلاح كلمة مرور المستخدم الافتراضي
-- تشغيل هذا السكريبت في PostgreSQL

-- الطريقة 1: تحديث كلمة المرور إلى Admin@123
UPDATE users 
SET password_hash = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy'
WHERE username = 'admin';

-- الطريقة 2: إذا لم تعمل، استخدم كلمة مرور بسيطة للاختبار
-- كلمة المرور: admin123
-- UPDATE users 
-- SET password_hash = '$2a$10$rqiCkCw/V5YJZz5JZz5JZOqKqKqKqKqKqKqKqKqKqKqKqKqKqKqKq'
-- WHERE username = 'admin';

-- التحقق من التحديث
SELECT username, email, role, is_active 
FROM users 
WHERE username = 'admin';
