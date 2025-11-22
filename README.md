# نظام ERP المالي لمنظمات المجتمع المدني
## CSO Financial ERP System

نظام مالي متكامل مصمم خصيصاً لمنظمات المجتمع المدني مع دعم كامل للأرصدة السالبة (العجز المؤقت).

---

## 📋 المتطلبات الأساسية

### قاعدة البيانات
- **PostgreSQL** 12 أو أحدث
- يجب أن يكون PostgreSQL مثبتاً ويعمل على جهازك

### Backend (Go)
- **Go** 1.21 أو أحدث
- يمكن تحميله من: https://golang.org/dl/

### Frontend (Node.js)
- **Node.js** 18 أو أحدث
- **npm** أو **yarn**
- يمكن تحميله من: https://nodejs.org/

---

## 🚀 خطوات التشغيل

### 1. إعداد قاعدة البيانات

#### أ. إنشاء قاعدة البيانات
افتح PostgreSQL وقم بتنفيذ:

```sql
CREATE DATABASE cso_erp;
```

#### ب. تنفيذ السكريبت الأساسي
قم بتنفيذ ملف `database/migrations/001_initial_schema.sql`:

**من خلال psql:**
```bash
psql -U postgres -d cso_erp -f database/migrations/001_initial_schema.sql
```

**أو من خلال pgAdmin:**
1. افتح pgAdmin
2. اتصل بقاعدة البيانات `cso_erp`
3. افتح Query Tool
4. انسخ محتوى ملف `001_initial_schema.sql` والصقه
5. اضغط Execute/Run

#### ج. التحقق من التثبيت
تحقق من إنشاء الجداول:
```sql
SELECT table_name FROM information_schema.tables 
WHERE table_schema = 'public';
```

يجب أن ترى الجداول التالية:
- users
- projects
- financial_containers
- transactions
- audit_log

---

### 2. إعداد Backend (Go)

#### أ. الانتقال إلى مجلد Backend
```bash
cd backend
```

#### ب. إنشاء ملف البيئة
انسخ ملف `.env.example` إلى `.env`:
```bash
copy .env.example .env
```

#### ج. تعديل إعدادات قاعدة البيانات
افتح ملف `.env` وعدّل الإعدادات:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=كلمة_المرور_الخاصة_بك
DB_NAME=cso_erp
DB_SSLMODE=disable

JWT_SECRET=غير-هذا-المفتاح-السري-في-الإنتاج
```

#### د. تحميل المكتبات
```bash
go mod download
```

#### هـ. تشغيل الخادم
```bash
go run cmd/server/main.go
```

يجب أن ترى:
```
✅ Database connected successfully
🚀 Server starting on port 8080
```

---

### 3. إعداد Frontend (Vue.js)

#### أ. الانتقال إلى مجلد Frontend
افتح نافذة Terminal جديدة:
```bash
cd frontend
```

#### ب. تثبيت المكتبات
```bash
npm install
```

#### ج. تشغيل خادم التطوير
```bash
npm run dev
```

يجب أن ترى:
```
  VITE v5.x.x  ready in xxx ms

  ➜  Local:   http://localhost:5173/
```

---

### 4. الوصول إلى النظام

افتح المتصفح واذهب إلى:
```
http://localhost:5173
```

#### بيانات الدخول الافتراضية:
- **اسم المستخدم:** `admin`
- **كلمة المرور:** `Admin@123`

> ⚠️ **مهم جداً:** يجب تغيير كلمة المرور بعد أول تسجيل دخول!

---

## 📁 هيكل المشروع

```
fac/
├── database/
│   └── migrations/
│       └── 001_initial_schema.sql    # سكريبت قاعدة البيانات
│
├── backend/                          # Go Backend
│   ├── cmd/
│   │   └── server/
│   │       └── main.go              # نقطة البداية
│   ├── internal/
│   │   ├── config/                  # إعدادات النظام
│   │   ├── database/                # اتصال قاعدة البيانات
│   │   ├── middleware/              # JWT, RBAC, Rate Limiting
│   │   ├── models/                  # نماذج البيانات
│   │   ├── services/                # منطق الأعمال
│   │   └── handlers/                # HTTP Handlers
│   ├── go.mod
│   └── .env                         # ملف الإعدادات (يجب إنشاؤه)
│
└── frontend/                         # Vue.js Frontend
    ├── src/
    │   ├── views/                   # صفحات النظام
    │   ├── components/              # مكونات قابلة لإعادة الاستخدام
    │   ├── stores/                  # Pinia State Management
    │   ├── router/                  # Vue Router
    │   └── api/                     # Axios API Client
    ├── package.json
    └── vite.config.js
```

---

## 🔑 الميزات الرئيسية

### 1. إدارة المقبوضات
- منح ومشاريع
- تبرعات واشتراكات
- تسجيل تلقائي للوجهة (نقدي/بنك)

### 2. إدارة المدفوعات
- العهدة المالية
- صرف مرتبات
- صرف تشغيلية
- **دعم الأرصدة السالبة** مع تحذير

### 3. التقارير
- الجرد اليومي الشامل
- تقرير حركة المشاريع
- تقرير الالتزامات والعجز

### 4. الأمان
- JWT Authentication مع HttpOnly Cookies
- Role-Based Access Control (RBAC)
- Rate Limiting
- Password Hashing (bcrypt)

---

## 👥 الصلاحيات

### Admin (مدير)
- صلاحيات كاملة
- يمكنه الحذف والتعديل

### Finance (مالي)
- إنشاء وتعديل العمليات
- لا يمكنه الحذف (فقط العكس)

### Auditor (مراجع)
- عرض التقارير فقط
- لا يمكنه التعديل

---

## 🔧 استكشاف الأخطاء

### مشكلة الاتصال بقاعدة البيانات
```
❌ Failed to connect to database
```
**الحل:**
1. تأكد من تشغيل PostgreSQL
2. تحقق من بيانات الاتصال في `.env`
3. تأكد من إنشاء قاعدة البيانات `cso_erp`

### خطأ في تشغيل Backend
```
go: cannot find module
```
**الحل:**
```bash
go mod tidy
go mod download
```

### خطأ في تشغيل Frontend
```
Module not found
```
**الحل:**
```bash
rm -rf node_modules package-lock.json
npm install
```

---

## 📞 الدعم الفني

للمساعدة أو الإبلاغ عن مشاكل، يرجى التواصل مع فريق التطوير.

---

## 📝 ملاحظات مهمة

1. **الأرصدة السالبة:** النظام يسمح بالعجز المؤقت مع تحذير واضح
2. **الأمان:** تأكد من تغيير `JWT_SECRET` في الإنتاج
3. **النسخ الاحتياطي:** قم بعمل نسخة احتياطية من قاعدة البيانات بانتظام
4. **التحديثات:** راجع ملف `CHANGELOG.md` لمعرفة التحديثات

---

## 🎯 الخطوات التالية

بعد التشغيل الناجح:

1. ✅ تغيير كلمة مرور المدير
2. ✅ إضافة المستخدمين
3. ✅ إنشاء المشاريع
4. ✅ تعريف الأوعية المالية (خزائن وبنوك)
5. ✅ البدء في تسجيل العمليات

---

**تم التطوير بواسطة:** فريق تطوير أنظمة منظمات المجتمع المدني  
**الإصدار:** 1.0.0  
**التاريخ:** 2024
