# 🔧 دليل الإصلاح النهائي - CSO ERP System

## المشكلة الحالية:
❌ `400 - POST /api/v1/transactions` - Invalid request body

## السبب:
Backend يتوقع حقول معينة لكن Frontend يرسل أسماء حقول مختلفة.

---

## ✅ الحل الكامل:

### الخطوة 1: إضافة بيانات أولية في قاعدة البيانات

افتح **pgAdmin** أو **psql** ونفذ:

```sql
-- الاتصال بقاعدة البيانات
\c cso-erp

-- إضافة صندوق نقدي
INSERT INTO financial_containers (name, name_ar, type, currency, balance, location, location_ar, is_active, created_by)
SELECT 
    'Main Cash Box',
    'الخزينة الرئيسية',
    'cash',
    'LYD',
    0.00,
    'Main Office',
    'المكتب الرئيسي',
    TRUE,
    (SELECT id FROM users WHERE role = 'admin' LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM financial_containers WHERE name = 'Main Cash Box');

-- إضافة حساب بنكي
INSERT INTO financial_containers (name, name_ar, type, currency, balance, bank_name, bank_name_ar, account_number, is_active, created_by)
SELECT
    'Al-Aman Bank Account',
    'حساب مصرف الأمان',
    'bank',
    'LYD',
    0.00,
    'Al-Aman Bank',
    'مصرف الأمان',
    '123456789',
    TRUE,
    (SELECT id FROM users WHERE role = 'admin' LIMIT 1)
WHERE NOT EXISTS (SELECT 1 FROM financial_containers WHERE name = 'Al-Aman Bank Account');

-- التحقق من البيانات
SELECT id, name, name_ar, type, currency, balance FROM financial_containers;
SELECT id, code, name, name_ar FROM projects;
```

---

### الخطوة 2: تصحيح أسماء الحقول في Frontend

المشكلة: Frontend يرسل `container_id` لكن Backend يتوقع `ContainerID` (PascalCase في JSON).

**الحل:** تأكد من أن أسماء الحقول في form تطابق ما يتوقعه Backend.

في ملف `GrantsView.vue`، تأكد من أن form يستخدم الأسماء الصحيحة.

---

### الخطوة 3: أعد تشغيل Backend

```bash
cd c:\xampp\htdocs\fac\backend
go run cmd\server\main.go
```

---

### الخطوة 4: اختبر النظام

1. افتح: `http://localhost:5173`
2. سجّل الدخول
3. اذهب إلى "منح ومشاريع"
4. املأ النموذج:
   - **المبلغ:** 1000
   - **وجهة الاستلام:** اختر "الخزينة الرئيسية"
   - **المشروع:** اختر "الإدارة العامة"
   - **البيان:** "منحة تجريبية"
5. اضغط "حفظ العملية"

---

## 🔍 تشخيص المشكلة:

إذا استمر الخطأ `400 - Invalid request body`، افحص:

### 1. في Browser Console (F12):
```javascript
// افتح Console وشاهد البيانات المرسلة
// يجب أن ترى شيئاً مثل:
{
  "type": "income_grant",
  "amount": 1000,
  "currency": "LYD",
  "exchange_rate": 1,
  "container_id": "uuid-here",
  "project_id": "uuid-here",
  "description": "منحة تجريبية"
}
```

### 2. تحقق من Backend Handler:
المشكلة قد تكون في أن Fiber لا يمكنه parse الـ JSON بشكل صحيح.

---

## 📝 الحل البديل - استخدام Console للتشخيص:

أضف هذا في `transaction_handler.go` بعد السطر 24:

```go
func (h *TransactionHandler) CreateTransaction(c *fiber.Ctx) error {
	var req models.CreateTransactionRequest
	
	// طباعة البيانات المستلمة للتشخيص
	body := c.Body()
	fmt.Printf("Received body: %s\n", string(body))
	
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}
	
	fmt.Printf("Parsed request: %+v\n", req)
	// ... بقية الكود
}
```

هذا سيطبع البيانات في Terminal ويساعدك على معرفة المشكلة بالضبط.

---

## ✅ الخلاصة:

1. ✅ أضف البيانات الأولية في قاعدة البيانات
2. ✅ تأكد من أن أسماء الحقول صحيحة
3. ✅ أعد تشغيل Backend
4. ✅ اختبر النظام

إذا استمرت المشكلة، أضف كود التشخيص أعلاه وأخبرني بما يظهر في Terminal!
