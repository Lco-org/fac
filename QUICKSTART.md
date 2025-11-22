# ملخص سريع - خطوات التشغيل

## 1️⃣ قاعدة البيانات
```bash
# في psql أو pgAdmin
CREATE DATABASE cso_erp;

# تنفيذ السكريبت
psql -U postgres -d cso_erp -f database\migrations\001_initial_schema.sql
```

## 2️⃣ Backend
```bash
cd backend
copy .env.example .env
# عدّل .env ووضع كلمة مرور PostgreSQL
go mod download
go run cmd\server\main.go
```

## 3️⃣ Frontend (نافذة جديدة)
```bash
cd frontend
npm install
npm run dev
```

## 4️⃣ الوصول
افتح: http://localhost:5173

**تسجيل الدخول:**
- Username: `admin`
- Password: `Admin@123`

---

## إنشاء بيانات تجريبية (اختياري)

```sql
-- خزينة نقدية
INSERT INTO financial_containers (name, name_ar, type, currency, balance, is_active, created_by)
VALUES ('Main Cash', 'الخزينة الرئيسية', 'cash', 'LYD', 0, TRUE, 
        (SELECT id FROM users WHERE username = 'admin'));

-- حساب بنكي
INSERT INTO financial_containers (name, name_ar, type, currency, balance, 
                                  bank_name, account_number, is_active, created_by)
VALUES ('Al-Aman Bank', 'مصرف الأمان', 'bank', 'LYD', 0, 
        'Al-Aman Bank', '123456', TRUE,
        (SELECT id FROM users WHERE username = 'admin'));
```

---

راجع **MANUAL_SETUP.md** للتفاصيل الكاملة
