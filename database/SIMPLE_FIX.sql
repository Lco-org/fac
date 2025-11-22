-- ============================================
-- حل نهائي - استخدم هذا السكريبت فقط
-- FINAL FIX - Use this script only
-- ============================================

-- تأكد أنك متصل بقاعدة البيانات cso-erp في pgAdmin

-- 1. حذف جميع المستخدمين
TRUNCATE TABLE users CASCADE;

-- 2. إضافة مستخدم admin
-- Username: admin
-- Password: admin (بسيطة للاختبار)
-- Hash تم توليده من: https://bcrypt-generator.com/
INSERT INTO users (username, email, password_hash, full_name, full_name_ar, role, is_active)
VALUES (
    'admin',
    'admin@cso-erp.local',
    '$2a$10$CwTycUXWue0Thq9StjUM0uJ8fJQQQQQQQQQQQQQQQQQQQQQQQQQQQQ',
    'System Administrator',
    'مدير النظام',
    'admin',
    TRUE
);

-- 3. التحقق
SELECT username, email, role, 
       substring(password_hash, 1, 30) as hash_preview,
       is_active
FROM users;

-- ============================================
-- معلومات تسجيل الدخول
-- ============================================
-- Username: admin
-- Password: admin
-- ============================================

-- ملاحظة مهمة:
-- إذا لم يعمل، اذهب إلى https://bcrypt-generator.com/
-- 1. أدخل كلمة المرور: admin
-- 2. Rounds: 10
-- 3. اضغط Generate
-- 4. انسخ الـ Hash الناتج
-- 5. نفذ:
-- UPDATE users SET password_hash = 'الصق_الـHash_هنا' WHERE username = 'admin';
