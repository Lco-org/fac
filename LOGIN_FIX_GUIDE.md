# 🔐 حل نهائي لمشكلة تسجيل الدخول

## المشكلة:
Backend يرجع **401 Unauthorized** لأن كلمة المرور المشفرة في قاعدة البيانات غير صحيحة.

---

## ✅ الحل النهائي:

### الخطوة 1: توليد Hash صحيح لكلمة المرور

اذهب إلى: **https://bcrypt-generator.com/**

#### لكلمة المرور `Admin@123`:
1. أدخل: `Admin@123`
2. Rounds: `10`
3. اضغط **Generate**
4. انسخ الـ Hash

#### لكلمة المرور `123456`:
1. أدخل: `123456`
2. Rounds: `10`
3. اضغط **Generate**
4. انسخ الـ Hash

---

### الخطوة 2: تحديث قاعدة البيانات

افتح **pgAdmin** أو **psql** ونفذ:

```sql
\c cso-erp

-- حذف المستخدمين القدامى
DELETE FROM users;

-- إضافة admin بـ Hash الجديد
INSERT INTO users (username, email, password_hash, full_name, full_name_ar, role, is_active)
VALUES (
    'admin',
    'admin@cso-erp.local',
    'الصق_هنا_الـHash_الذي_حصلت_عليه_من_bcrypt-generator',
    'System Administrator',
    'مدير النظام',
    'admin',
    TRUE
);

-- إضافة Ezy بـ Hash الجديد
INSERT INTO users (username, email, password_hash, full_name, full_name_ar, role, is_active)
VALUES (
    'Ezy',
    'ezy@cso-erp.local',
    'الصق_هنا_الـHash_الذي_حصلت_عليه_من_bcrypt-generator',
    'Ezy User',
    'المستخدم إيزي',
    'admin',
    TRUE
);

-- التحقق
SELECT username, email, role FROM users;
```

---

### الخطوة 3: أعد تشغيل Backend

اضغط `Ctrl+C` في نافذة Backend ثم:

```bash
cd c:\xampp\htdocs\fac\backend
go run cmd\server\main.go
```

---

### الخطوة 4: سجّل الدخول

افتح المتصفح: **http://localhost:5173**

استخدم:
- **Username:** `admin`
- **Password:** `Admin@123`

أو:
- **Username:** `Ezy`
- **Password:** `123456`

---

## 🎯 مثال عملي:

إذا كان الـ Hash الذي حصلت عليه من bcrypt-generator هو:
```
$2a$10$abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJ
```

نفذ:
```sql
UPDATE users 
SET password_hash = '$2a$10$abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJ'
WHERE username = 'admin';
```

---

**يجب أن يعمل تسجيل الدخول الآن!** ✅
