# 🔧 دليل حل المشاكل النهائي

## المشكلة الحالية:
- Frontend لا يعمل (localhost:5173 لا يفتح)
- Backend يعمل ✅ لكن لا فائدة بدون Frontend

---

## ✅ الحل الكامل خطوة بخطوة:

### الخطوة 1: أوقف جميع العمليات

اضغط `Ctrl+C` في جميع نوافذ Terminal المفتوحة.

---

### الخطوة 2: تأكد من قاعدة البيانات

افتح pgAdmin أو psql:

```sql
-- التحقق من وجود قاعدة البيانات
SELECT datname FROM pg_database WHERE datname = 'cso-erp';

-- إذا لم تكن موجودة، أنشئها:
CREATE DATABASE "cso-erp";

-- الاتصال بها
\c cso-erp

-- تنفيذ السكريبت (إذا لم يتم تنفيذه)
-- من psql:
-- \i c:/xampp/htdocs/fac/database/migrations/001_initial_schema.sql

-- التحقق من المستخدمين
SELECT username, email, role FROM users;
```

**يجب أن ترى:**
- `admin` | admin@cso-erp.local | admin
- `Ezy` | ezy@cso-erp.local | admin

---

### الخطوة 3: تشغيل Backend

**نافذة PowerShell الأولى:**

```bash
cd c:\xampp\htdocs\fac\backend
go run cmd\server\main.go
```

**يجب أن ترى:**
```
✅ Database connected successfully
🚀 Server starting on port 8080
📊 Environment: development
🔒 CORS Origin: http://localhost:5173

┌───────────────────────────────────────────────────┐
│                  CSO ERP System                   │
│                   Fiber v2.52.0                   │
│               http://127.0.0.1:8080               │
└───────────────────────────────────────────────────┘
```

**⚠️ اترك هذه النافذة مفتوحة!**

---

### الخطوة 4: تشغيل Frontend

**نافذة PowerShell الثانية:**

```bash
cd c:\xampp\htdocs\fac\frontend

# إذا كانت هناك مشاكل، احذف node_modules وأعد التثبيت:
rm -rf node_modules package-lock.json
npm install

# ثم شغّل:
npm run dev
```

**يجب أن ترى:**
```
VITE v5.4.21  ready in xxx ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
```

**⚠️ اترك هذه النافذة أيضاً مفتوحة!**

---

### الخطوة 5: افتح المتصفح

اذهب إلى: **http://localhost:5173**

---

### الخطوة 6: سجّل الدخول

**الخيار 1:**
- Username: `admin`
- Password: `Admin@123`

**الخيار 2:**
- Username: `Ezy`
- Password: `123456`

---

## 🔍 إذا استمرت المشاكل:

### مشكلة: Frontend لا يبدأ

```bash
cd c:\xampp\htdocs\fac\frontend
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
npm run dev
```

### مشكلة: Backend لا يتصل بقاعدة البيانات

تحقق من ملف `backend\.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=MlP363722
DB_NAME=cso-erp
DB_SSLMODE=disable
```

### مشكلة: تسجيل الدخول يفشل

نفذ في PostgreSQL:
```sql
\c cso-erp

-- تحديث كلمات المرور
UPDATE users 
SET password_hash = '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'
WHERE username = 'admin';

INSERT INTO users (username, email, password_hash, full_name, full_name_ar, role, is_active)
VALUES ('Ezy', 'ezy@cso-erp.local', 
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        'Ezy User', 'المستخدم إيزي', 'admin', TRUE)
ON CONFLICT (username) DO UPDATE SET password_hash = EXCLUDED.password_hash;
```

---

## 📋 قائمة التحقق النهائية:

- [ ] PostgreSQL يعمل
- [ ] قاعدة البيانات `cso-erp` موجودة
- [ ] الجداول والمستخدمين موجودين
- [ ] Backend يعمل على المنفذ 8080
- [ ] Frontend يعمل على المنفذ 5173
- [ ] يمكن فتح http://localhost:5173
- [ ] يمكن تسجيل الدخول

---

**إذا اتبعت هذه الخطوات بالترتيب، يجب أن يعمل كل شيء! ✅**
