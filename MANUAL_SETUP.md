# دليل التنفيذ اليدوي - نظام ERP المالي
## خطوات تشغيل المشروع بالتفصيل

---

## المرحلة 1️⃣: إعداد قاعدة البيانات PostgreSQL

### الخطوة 1: إنشاء قاعدة البيانات

**افتح pgAdmin أو psql وقم بتنفيذ:**

```sql
CREATE DATABASE cso_erp;
```

### الخطوة 2: تنفيذ السكريبت

**الطريقة الأولى - من خلال psql:**
```bash
cd c:\xampp\htdocs\fac
psql -U postgres -d cso_erp -f database\migrations\001_initial_schema.sql
```

**الطريقة الثانية - من خلال pgAdmin:**
1. افتح pgAdmin
2. اتصل بقاعدة البيانات `cso_erp`
3. انقر بزر الماوس الأيمن على قاعدة البيانات → Query Tool
4. افتح ملف `database\migrations\001_initial_schema.sql`
5. انسخ المحتوى كاملاً والصقه في Query Tool
6. اضغط F5 أو زر Execute

### الخطوة 3: التحقق من نجاح التنفيذ

```sql
-- تحقق من الجداول
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public'
ORDER BY table_name;

-- يجب أن ترى:
-- audit_log
-- financial_containers
-- projects
-- transactions
-- users

-- تحقق من المستخدم الافتراضي
SELECT username, email, role FROM users;
-- يجب أن ترى: admin | admin@cso-erp.local | admin
```

✅ **إذا ظهرت النتائج بنجاح، انتقل للمرحلة التالية**

---

## المرحلة 2️⃣: إعداد وتشغيل Backend (Go)

### الخطوة 1: التأكد من تثبيت Go

افتح PowerShell أو CMD وتحقق:
```bash
go version
```
يجب أن ترى: `go version go1.21.x` أو أحدث

**إذا لم يكن مثبتاً:**
- حمّل من: https://golang.org/dl/
- ثبّت وأعد تشغيل Terminal

### الخطوة 2: إنشاء ملف الإعدادات

```bash
cd c:\xampp\htdocs\fac\backend
copy .env.example .env
```

### الخطوة 3: تعديل ملف .env

افتح ملف `backend\.env` في أي محرر نصوص وعدّل:

```env
# إعدادات قاعدة البيانات
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=كلمة_المرور_الخاصة_بك_هنا
DB_NAME=cso_erp
DB_SSLMODE=disable

# إعدادات JWT (غيّر السر في الإنتاج)
JWT_SECRET=my-super-secret-key-change-in-production-2024
JWT_EXPIRY=24h

# باقي الإعدادات (اتركها كما هي)
APP_ENV=development
APP_PORT=8080
CORS_ORIGIN=http://localhost:5173
```

⚠️ **مهم:** تأكد من وضع كلمة مرور PostgreSQL الصحيحة!

### الخطوة 4: تحميل المكتبات

```bash
cd c:\xampp\htdocs\fac\backend
go mod tidy
go mod download
```

**الانتظار:** قد يستغرق هذا 2-5 دقائق حسب سرعة الإنترنت

### الخطوة 5: تشغيل الخادم

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
```

✅ **إذا ظهرت هذه الرسائل، Backend يعمل بنجاح!**

❌ **إذا ظهر خطأ في الاتصال بقاعدة البيانات:**
- تحقق من تشغيل PostgreSQL
- راجع بيانات الاتصال في `.env`
- تأكد من كلمة المرور

**اترك هذه النافذة مفتوحة وتعمل!**

---

## المرحلة 3️⃣: إعداد وتشغيل Frontend (Vue.js)

### الخطوة 1: التأكد من تثبيت Node.js

افتح نافذة PowerShell **جديدة** وتحقق:
```bash
node --version
npm --version
```

يجب أن ترى:
- Node: `v18.x.x` أو أحدث
- npm: `9.x.x` أو أحدث

**إذا لم يكن مثبتاً:**
- حمّل من: https://nodejs.org/
- اختر LTS version
- ثبّت وأعد تشغيل Terminal

### الخطوة 2: تثبيت المكتبات

```bash
cd c:\xampp\htdocs\fac\frontend
npm install
```

**الانتظار:** قد يستغرق هذا 3-7 دقائق

**إذا ظهرت تحذيرات (warnings):** لا بأس، المهم عدم وجود أخطاء (errors)

### الخطوة 3: تشغيل خادم التطوير

```bash
cd c:\xampp\htdocs\fac\frontend
npm run dev
```

**يجب أن ترى:**
```
  VITE v5.x.x  ready in xxx ms

  ➜  Local:   http://localhost:5173/
  ➜  Network: use --host to expose
  ➜  press h + enter to show help
```

✅ **Frontend يعمل بنجاح!**

**اترك هذه النافذة أيضاً مفتوحة وتعمل!**

---

## المرحلة 4️⃣: الوصول إلى النظام

### الخطوة 1: فتح المتصفح

افتح أي متصفح (Chrome, Firefox, Edge) واذهب إلى:
```
http://localhost:5173
```

### الخطوة 2: تسجيل الدخول

**بيانات الدخول الافتراضية:**
- **اسم المستخدم:** `admin`
- **كلمة المرور:** `Admin@123`

### الخطوة 3: استكشاف النظام

بعد تسجيل الدخول، يجب أن ترى:
- ✅ لوحة التحكم (Dashboard)
- ✅ القائمة الجانبية مع الخيارات
- ✅ ملخص الموقف المالي

---

## المرحلة 5️⃣: اختبار الوظائف الأساسية

### اختبار 1: عرض التقارير

1. من القائمة الجانبية → **التقارير**
2. اختر **الجرد اليومي**
3. يجب أن ترى جدول فارغ (لا توجد بيانات بعد)

### اختبار 2: إنشاء عملية مالية (اختياري)

⚠️ **ملاحظة:** لن تعمل العمليات المالية حالياً لأننا لم ننشئ:
- خزائن (Containers)
- مشاريع إضافية

**لإنشاء بيانات تجريبية، نفذ في PostgreSQL:**

```sql
-- إنشاء خزينة نقدية
INSERT INTO financial_containers (name, name_ar, type, currency, balance, is_active, created_by)
VALUES (
    'Main Cash Box',
    'الخزينة الرئيسية',
    'cash',
    'LYD',
    0.00,
    TRUE,
    (SELECT id FROM users WHERE username = 'admin')
);

-- إنشاء حساب بنكي
INSERT INTO financial_containers (name, name_ar, type, currency, balance, bank_name, bank_name_ar, account_number, is_active, created_by)
VALUES (
    'Al-Aman Bank',
    'مصرف الأمان',
    'bank',
    'LYD',
    0.00,
    'Al-Aman Bank',
    'مصرف الأمان',
    '123456789',
    TRUE,
    (SELECT id FROM users WHERE username = 'admin')
);
```

الآن يمكنك تجربة إنشاء عملية صرف أو قبض!

---

## 🔧 استكشاف الأخطاء الشائعة

### مشكلة: Backend لا يعمل

**الخطأ:** `Failed to connect to database`

**الحل:**
1. تأكد من تشغيل PostgreSQL
2. افتح pgAdmin وتحقق من الاتصال
3. راجع معلومات `.env`
4. جرب الاتصال يدوياً:
   ```bash
   psql -U postgres -d cso_erp
   ```

### مشكلة: Frontend لا يعمل

**الخطأ:** `Module not found`

**الحل:**
```bash
cd c:\xampp\htdocs\fac\frontend
rm -rf node_modules package-lock.json
npm install
```

### مشكلة: CORS Error في المتصفح

**الحل:**
- تأكد من تشغيل Backend على المنفذ 8080
- تأكد من تشغيل Frontend على المنفذ 5173
- أعد تشغيل كلا الخادمين

### مشكلة: المنفذ مستخدم

**الخطأ:** `Port 8080 already in use`

**الحل:**
```bash
# ابحث عن البرنامج المستخدم للمنفذ
netstat -ano | findstr :8080

# أوقف البرنامج أو غيّر المنفذ في .env
```

---

## 📊 ملخص الحالة

عند التشغيل الناجح، يجب أن يكون لديك:

✅ **3 نوافذ Terminal مفتوحة:**
1. PostgreSQL (يعمل في الخلفية)
2. Backend (Go) - يعمل على المنفذ 8080
3. Frontend (Vite) - يعمل على المنفذ 5173

✅ **متصفح مفتوح على:**
- http://localhost:5173

✅ **قاعدة بيانات:**
- تحتوي على الجداول والبيانات الأولية

---

## 🎯 الخطوات التالية

بعد التشغيل الناجح:

1. ✅ **غيّر كلمة مرور المدير** (مهم جداً!)
2. ✅ أضف مستخدمين جدد
3. ✅ أنشئ المشاريع
4. ✅ أنشئ الخزائن والحسابات البنكية
5. ✅ ابدأ تسجيل العمليات المالية

---

## 📞 في حالة وجود مشاكل

إذا واجهت أي مشكلة:

1. تحقق من رسائل الخطأ في Terminal
2. راجع ملف `.env`
3. تأكد من تشغيل PostgreSQL
4. تحقق من المنافذ المستخدمة

---

**تم إعداد هذا الدليل بواسطة:** نظام CSO ERP  
**التاريخ:** 2024  
**الإصدار:** 1.0
