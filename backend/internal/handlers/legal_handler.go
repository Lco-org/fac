package handlers

import (
	"github.com/gofiber/fiber/v2"
)

// LegalHandler handles legal pages (Privacy Policy, Terms of Service)
type LegalHandler struct{}

// NewLegalHandler creates a new legal handler
func NewLegalHandler() *LegalHandler {
	return &LegalHandler{}
}

// GetPrivacyPolicy returns the privacy policy content
func (h *LegalHandler) GetPrivacyPolicy(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"title":        "سياسة الخصوصية",
		"title_en":     "Privacy Policy",
		"last_updated": "2024-01-01",
		"content": `
# سياسة الخصوصية

## مقدمة
نحن في نظام CSO-ERP نلتزم بحماية خصوصيتك. توضح هذه السياسة كيفية جمع واستخدام وحماية معلوماتك الشخصية.

## المعلومات التي نجمعها
- **معلومات الحساب**: الاسم، البريد الإلكتروني، رقم الهاتف
- **معلومات المالية**: المعاملات، الأرصدة، التقارير
- **معلومات الاستخدام**: سجلات الدخول، النشاط على النظام

## كيف نستخدم المعلومات
- تقديم الخدمات المطلوبة
- تحسين تجربة المستخدم
- إرسال إشعارات مهمة
- الامتثال للمتطلبات القانونية

## حماية المعلومات
- تشفير البيانات أثناء النقل والتخزين
- تقييد الوصول للموظفين المصرح لهم
- مراجعات أمنية دورية
- نسخ احتياطية منتظمة

## حقوقك
- الوصول إلى بياناتك
- تصحيح البيانات غير الدقيقة
- طلب حذف البيانات
- الاعتراض على المعالجة

## ملفات تعريف الارتباط (Cookies)
نستخدم ملفات الـ Cookies لـ:
- الحفاظ على جلسة تسجيل الدخول
- تذكر تفضيلاتك
- تحليل استخدام النظام

## الاتصال بنا
للاستفسارات حول الخصوصية:
- البريد الإلكتروني: privacy@lco.org.ly
- الهاتف: +218-XX-XXXXXXX
`,
		"content_en": `
# Privacy Policy

## Introduction
At CSO-ERP, we are committed to protecting your privacy. This policy explains how we collect, use, and protect your personal information.

## Information We Collect
- **Account Information**: Name, email, phone number
- **Financial Information**: Transactions, balances, reports
- **Usage Information**: Login logs, system activity

## How We Use Information
- Providing requested services
- Improving user experience
- Sending important notifications
- Compliance with legal requirements

## Information Protection
- Encryption of data in transit and at rest
- Access restricted to authorized personnel
- Regular security audits
- Regular backups

## Your Rights
- Access your data
- Correct inaccurate data
- Request data deletion
- Object to processing

## Cookies
We use cookies for:
- Maintaining login sessions
- Remembering your preferences
- Analyzing system usage

## Contact Us
For privacy inquiries:
- Email: privacy@lco.org.ly
- Phone: +218-XX-XXXXXXX
`,
	})
}

// GetTermsOfService returns the terms of service content
func (h *LegalHandler) GetTermsOfService(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"title":        "شروط الخدمة",
		"title_en":     "Terms of Service",
		"last_updated": "2024-01-01",
		"content": `
# شروط الخدمة

## 1. قبول الشروط
باستخدامك لنظام CSO-ERP، فإنك توافق على هذه الشروط والأحكام.

## 2. وصف الخدمة
نظام CSO-ERP هو منصة سحابية لإدارة الموارد المالية والإدارية لمنظمات المجتمع المدني.

## 3. حسابات المستخدمين
- يجب تقديم معلومات دقيقة عند التسجيل
- أنت مسؤول عن الحفاظ على سرية كلمة المرور
- يجب إبلاغنا فوراً بأي استخدام غير مصرح به

## 4. الاستخدام المقبول
يُحظر استخدام النظام في:
- أي نشاط غير قانوني
- محاولة الوصول غير المصرح به
- نشر محتوى ضار أو مسيء
- التدخل في عمل النظام

## 5. الملكية الفكرية
- جميع حقوق النظام محفوظة
- لا يجوز نسخ أو توزيع البرنامج
- أنت تحتفظ بملكية بياناتك

## 6. حدود المسؤولية
- النظام مقدم "كما هو"
- لا نضمن الخلو من الأخطاء
- لسنا مسؤولين عن الأضرار غير المباشرة

## 7. الإنهاء
يحق لنا إنهاء أو تعليق حسابك في حالة:
- انتهاك هذه الشروط
- عدم الدفع
- طلبك إلغاء الاشتراك

## 8. التعديلات
قد نقوم بتحديث هذه الشروط. سيتم إخطارك بالتغييرات الجوهرية.

## 9. القانون الحاكم
تخضع هذه الشروط للقوانين الليبية.

## 10. الاتصال
- البريد الإلكتروني: legal@lco.org.ly
- الهاتف: +218-XX-XXXXXXX
`,
		"content_en": `
# Terms of Service

## 1. Acceptance of Terms
By using CSO-ERP, you agree to these terms and conditions.

## 2. Service Description
CSO-ERP is a cloud platform for managing financial and administrative resources for civil society organizations.

## 3. User Accounts
- You must provide accurate information when registering
- You are responsible for maintaining password confidentiality
- You must notify us immediately of any unauthorized use

## 4. Acceptable Use
The system may not be used for:
- Any illegal activity
- Unauthorized access attempts
- Spreading harmful or offensive content
- Interfering with system operation

## 5. Intellectual Property
- All system rights are reserved
- Software may not be copied or distributed
- You retain ownership of your data

## 6. Limitation of Liability
- The system is provided "as is"
- We do not guarantee error-free operation
- We are not liable for indirect damages

## 7. Termination
We may terminate or suspend your account if:
- You violate these terms
- Non-payment
- You request cancellation

## 8. Modifications
We may update these terms. You will be notified of material changes.

## 9. Governing Law
These terms are governed by Libyan law.

## 10. Contact
- Email: legal@lco.org.ly
- Phone: +218-XX-XXXXXXX
`,
	})
}
