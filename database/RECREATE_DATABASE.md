# 🔄 إعادة إنشاء قاعدة البيانات من الصفر

## ⚠️ تحذير
هذا سيحذف جميع البيانات الموجودة! استخدم فقط إذا كنت تريد البدء من جديد.

---

## الخطوات:

### 1. حذف قاعدة البيانات القديمة

افتح pgAdmin أو psql ونفذ:

```sql
-- قطع جميع الاتصالات
SELECT pg_terminate_backend(pg_stat_activity.pid)
FROM pg_stat_activity
WHERE pg_stat_activity.datname = 'cso_erp'
  AND pid <> pg_backend_pid();

-- حذف قاعدة البيانات
DROP DATABASE IF EXISTS cso_erp;
```

---

### 2. إنشاء قاعدة بيانات جديدة

```sql
CREATE DATABASE cso_erp
    WITH 
    OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TABLESPACE = pg_default
    CONNECTION LIMIT = -1;
```

---

### 3. تنفيذ السكريبت الجديد

**من خلال psql:**
```bash
psql -U postgres -d cso_erp -f c:\xampp\htdocs\fac\database\migrations\001_initial_schema.sql
```

**من خلال pgAdmin:**
1. افتح pgAdmin
2. اتصل بقاعدة البيانات `cso_erp`
3. Query Tool (F5)
4. افتح ملف `001_initial_schema.sql`
5. Execute (F5)

---

### 4. التحقق من النجاح

```sql
-- التحقق من الجداول
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public'
ORDER BY table_name;

-- التحقق من المستخدمين
SELECT username, email, role, is_active 
FROM users;

-- يجب أن ترى:
-- admin | admin@cso-erp.local | admin | true
-- Ezy   | ezy@cso-erp.local   | admin | true
```

---

### 5. أعد تشغيل Backend

```bash
cd c:\xampp\htdocs\fac\backend
go run cmd\server\main.go
```

---

### 6. سجّل الدخول

افتح المتصفح: **http://localhost:5173**

**الخيار 1:**
- Username: `admin`
- Password: `Admin@123`

**الخيار 2:**
- Username: `Ezy`
- Password: `123456`

---

## ✅ يجب أن يعمل الآن بشكل صحيح!

كلمات المرور مشفرة بشكل صحيح باستخدام bcrypt.
