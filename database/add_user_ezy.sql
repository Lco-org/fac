-- إضافة مستخدم جديد: Ezy
-- كلمة المرور: 123456
-- الصلاحية: admin (يمكن تغييرها إلى 'finance' أو 'auditor')

-- حذف المستخدم إذا كان موجوداً (اختياري)
DELETE FROM users WHERE username = 'Ezy';

-- إضافة المستخدم الجديد
INSERT INTO users (
    username, 
    email, 
    password_hash, 
    full_name, 
    full_name_ar, 
    role, 
    is_active
)
VALUES (
    'Ezy',
    'ezy@cso-erp.local',
    '$2a$10$YQ7Y7Y7Y7Y7Y7Y7Y7Y7Y7OqKqKqKqKqKqKqKqKqKqKqKqKqKqKqKqK',
    'Ezy User',
    'المستخدم إيزي',
    'admin',
    TRUE
);

-- التحقق من إضافة المستخدم
SELECT username, email, role, is_active 
FROM users 
WHERE username = 'Ezy';

-- ملاحظة: hash كلمة المرور أعلاه هو مثال فقط
-- لإنشاء hash صحيح لكلمة المرور 123456، يجب استخدام bcrypt
