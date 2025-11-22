# حل المشاكل - Troubleshooting Guide

## ❌ مشكلة 1: Backend - خطأ كلمة المرور

**الخطأ:**
```
❌ Failed to connect to database: pq: password authentication failed for user "postgres"
```

**الحل:**

### الخطوة 1: تحديد كلمة مرور PostgreSQL الصحيحة

افتح pgAdmin وحاول الاتصال، أو جرب:
```bash
psql -U postgres
```

إذا طلب كلمة مرور، استخدمها في الخطوة التالية.

### الخطوة 2: تعديل ملف .env

افتح ملف `backend\.env` وعدّل السطر:
```env
DB_PASSWORD=ضع_كلمة_المرور_الصحيحة_هنا
```

**مثال:**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=123456
DB_NAME=cso_erp
DB_SSLMODE=disable
```

### الخطوة 3: أعد تشغيل Backend

```bash
cd c:\xampp\htdocs\fac\backend
go run cmd\server\main.go
```

يجب أن ترى:
```
✅ Database connected successfully
🚀 Server starting on port 8080
```

---

## ✅ مشكلة 2: Frontend - PostCSS Config (تم الحل)

**الخطأ:**
```
module is not defined in ES module scope
```

**الحل:** تم تحديث ملف `postcss.config.js` تلقائياً

### الآن أعد تشغيل Frontend:

```bash
cd c:\xampp\htdocs\fac\frontend
npm run dev
```

يجب أن يعمل بدون مشاكل! ✅

---

## 🎯 الخطوات الكاملة للتشغيل الآن:

### 1. Backend (نافذة Terminal أولى):
```bash
cd c:\xampp\htdocs\fac\backend
# تأكد من تعديل .env بكلمة المرور الصحيحة
go run cmd\server\main.go
```

### 2. Frontend (نافذة Terminal ثانية):
```bash
cd c:\xampp\htdocs\fac\frontend
npm run dev
```

### 3. افتح المتصفح:
```
http://localhost:5173
```

**تسجيل الدخول:**
- Username: `admin`
- Password: `Admin@123`

---

## 🔍 كيف تعرف كلمة مرور PostgreSQL؟

### الطريقة 1: من خلال pgAdmin
1. افتح pgAdmin
2. عند الاتصال، ستطلب كلمة المرور
3. استخدم نفس الكلمة في `.env`

### الطريقة 2: إعادة تعيين كلمة المرور (إذا نسيتها)

**في Windows:**

1. ابحث عن ملف `pg_hba.conf`:
   - عادة في: `C:\Program Files\PostgreSQL\15\data\pg_hba.conf`

2. افتحه كـ Administrator وغيّر السطر:
   ```
   host    all             all             127.0.0.1/32            scram-sha-256
   ```
   إلى:
   ```
   host    all             all             127.0.0.1/32            trust
   ```

3. أعد تشغيل PostgreSQL Service

4. الآن يمكنك الاتصال بدون كلمة مرور:
   ```bash
   psql -U postgres
   ALTER USER postgres PASSWORD 'new_password_here';
   ```

5. أرجع التغيير في `pg_hba.conf` إلى `scram-sha-256`

6. أعد تشغيل PostgreSQL Service مرة أخرى

---

## ✅ التحقق من نجاح التشغيل

عند التشغيل الناجح:

**Backend يجب أن يظهر:**
```
✅ Database connected successfully
🚀 Server starting on port 8080
📊 Environment: development
🔒 CORS Origin: http://localhost:5173
```

**Frontend يجب أن يظهر:**
```
VITE v5.4.21  ready in xxx ms

➜  Local:   http://localhost:5173/
➜  Network: use --host to expose
```

**بدون أي أخطاء!** ✅
