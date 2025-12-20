package services

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

type EmailConfig struct {
	SMTPHost    string
	SMTPPort    string
	Username    string
	Password    string
	FromAddress string
	FromName    string
	UseTLS      bool
	SkipVerify  bool
}

type EmailService struct {
	config *EmailConfig
	mock   bool
}

const (
	MaxEmailSize       = 10 * 1024 * 1024
	SMTPDialTimeout    = 30 * time.Second
	MaxEmailsPerMinute = 10
	MinTLSVersion      = tls.VersionTLS12
)

// EmailRecipient represents a recipient for bulk emails
type EmailRecipient struct {
	ID    string
	Email string
	Name  string
}

// BulkEmailResult contains the results of a bulk email operation
type BulkEmailResult struct {
	Total     int      `json:"total"`
	Sent      int      `json:"sent"`
	Failed    int      `json:"failed"`
	FailedIDs []string `json:"failed_ids,omitempty"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func NewEmailService(config *EmailConfig) *EmailService {
	if config == nil || config.SMTPHost == "" {
		log.Println("Email service running in mock mode")
		return &EmailService{config: nil, mock: true}
	}

	if !emailRegex.MatchString(config.FromAddress) {
		log.Printf("Invalid FROM address: %s - defaulting to mock mode", config.FromAddress)
		return &EmailService{config: nil, mock: true}
	}

	if config.SkipVerify {
		log.Println("SECURITY WARNING: TLS certificate verification is DISABLED")
	}

	return &EmailService{config: config, mock: false}
}

func NewMockEmailService() *EmailService {
	return &EmailService{mock: true}
}

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("empty email")
	}
	if len(email) > 254 {
		return fmt.Errorf("email too long")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid format")
	}
	if strings.ContainsAny(email, "\r\n") {
		return fmt.Errorf("invalid characters")
	}
	return nil
}

func sanitizeHeader(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	if len(value) > 998 {
		value = value[:998]
	}
	return value
}

func (s *EmailService) SendOTPEmail(ctx context.Context, toEmail, toName, code string) error {
	if err := validateEmail(toEmail); err != nil {
		return err
	}

	toName = sanitizeHeader(toName)

	if len(code) != 6 {
		return fmt.Errorf("invalid OTP length")
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return fmt.Errorf("invalid OTP format")
		}
	}

	subject := "رمز التحقق - CSO ERP"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1a56db; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .code-box { background: #1a56db; color: white; font-size: 32px; letter-spacing: 8px; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px; font-family: monospace; }
        .warning { color: #dc2626; font-size: 14px; margin-top: 20px; padding: 10px; background: #fef2f2; border-radius: 4px; }
        .footer { background: #374151; color: #9ca3af; padding: 15px; text-align: center; font-size: 12px; border-radius: 0 0 8px 8px; }
        .security-note { font-size: 11px; color: #6b7280; margin-top: 15px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>نظام إدارة الموارد المالية</h1>
        </div>
        <div class="content">
            <p>مرحباً <strong>%s</strong>،</p>
            <p>تم طلب تسجيل الدخول إلى حسابك. استخدم رمز التحقق التالي:</p>
            <div class="code-box">%s</div>
            <p>صلاحية هذا الرمز: <strong>10 دقائق</strong></p>
            <div class="warning">
                ⚠️ <strong>تحذير أمني:</strong> إذا لم تطلب تسجيل الدخول، يرجى تجاهل هذا البريد وتغيير كلمة المرور فوراً.
                <br>لا تشارك هذا الرمز مع أي شخص. موظفونا لن يطلبوا منك هذا الرمز أبداً.
            </div>
            <p class="security-note">
                🔒 هذا البريد مرسل من نظام CSO ERP الآمن. تم إرسال هذه الرسالة في %s
            </p>
        </div>
        <div class="footer">
            <p>هذا بريد آلي، يرجى عدم الرد عليه.</p>
            <p>© منظمة LCO - جميع الحقوق محفوظة</p>
        </div>
    </div>
</body>
</html>`, toName, code, time.Now().Format("2006-01-02 15:04:05"))

	return s.SendEmail(ctx, toEmail, subject, body, true)
}

// SendSSA2FACode sends 2FA code for Super System Admin login
func (s *EmailService) SendSSA2FACode(ctx context.Context, toEmail, toName, code string) error {
	if err := validateEmail(toEmail); err != nil {
		return err
	}

	toName = sanitizeHeader(toName)

	subject := "🔐 رمز التحقق - لوحة تحكم مشرف النظام"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; margin: 0; padding: 0; background-color: #0f172a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #7c3aed 0%%, #4f46e5 100%%); color: white; padding: 30px; text-align: center; border-radius: 12px 12px 0 0; }
        .header h1 { margin: 0; font-size: 24px; }
        .header .badge { display: inline-block; background: rgba(255,255,255,0.2); padding: 5px 15px; border-radius: 20px; font-size: 12px; margin-top: 10px; }
        .content { background: #1e293b; padding: 30px; border: 1px solid #334155; color: #e2e8f0; }
        .code-box { background: linear-gradient(135deg, #7c3aed 0%%, #4f46e5 100%%); color: white; font-size: 36px; letter-spacing: 10px; padding: 25px; text-align: center; margin: 25px 0; border-radius: 12px; font-family: 'Courier New', monospace; font-weight: bold; box-shadow: 0 10px 30px rgba(124, 58, 237, 0.3); }
        .warning { color: #fbbf24; font-size: 14px; margin-top: 25px; padding: 15px; background: rgba(251, 191, 36, 0.1); border-radius: 8px; border-right: 4px solid #fbbf24; }
        .security-info { background: rgba(239, 68, 68, 0.1); border: 1px solid rgba(239, 68, 68, 0.3); border-radius: 8px; padding: 15px; margin-top: 20px; }
        .security-info h4 { color: #ef4444; margin: 0 0 10px 0; }
        .footer { background: #0f172a; color: #64748b; padding: 20px; text-align: center; font-size: 12px; border-radius: 0 0 12px 12px; }
        .shield { font-size: 48px; margin-bottom: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="shield">🛡️</div>
            <h1>لوحة تحكم مشرف النظام</h1>
            <div class="badge">SUPER SYSTEM ADMIN</div>
        </div>
        <div class="content">
            <p>مرحباً <strong>%s</strong>،</p>
            <p>تم تسجيل محاولة دخول إلى لوحة تحكم مشرف النظام. استخدم رمز التحقق التالي:</p>
            <div class="code-box">%s</div>
            <p style="text-align: center; color: #94a3b8;">⏱️ صلاحية الرمز: <strong style="color: #22c55e;">10 دقائق</strong></p>
            
            <div class="warning">
                ⚠️ <strong>تنبيه أمني شديد:</strong><br>
                هذا الرمز يمنح وصولاً كاملاً لإدارة جميع المنظمات. إذا لم تكن أنت من طلب تسجيل الدخول، قم فوراً بما يلي:
                <ul style="margin: 10px 0; padding-right: 20px;">
                    <li>لا تشارك هذا الرمز مع أي شخص</li>
                    <li>تواصل مع فريق الأمان فوراً</li>
                    <li>قم بتغيير كلمة المرور</li>
                </ul>
            </div>
            
            <div class="security-info">
                <h4>🔒 معلومات الجلسة</h4>
                <p style="margin: 5px 0; font-size: 13px;">الوقت: %s</p>
                <p style="margin: 5px 0; font-size: 13px;">النوع: تسجيل دخول مشرف أعلى</p>
            </div>
        </div>
        <div class="footer">
            <p>هذا بريد آلي عالي الأمان من نظام CSO ERP</p>
            <p>© منظمة LCO - جميع الحقوق محفوظة</p>
        </div>
    </div>
</body>
</html>`, toName, code, time.Now().Format("2006-01-02 15:04:05"))

	return s.SendEmail(ctx, toEmail, subject, body, true)
}

// SendWelcomeEmail sends a welcome message to newly created users
func (s *EmailService) SendWelcomeEmail(ctx context.Context, toEmail, fullName, username, role string) error {
	if err := validateEmail(toEmail); err != nil {
		return err
	}

	fullName = sanitizeHeader(fullName)

	roleArabic := map[string]string{
		"admin":   "مدير النظام",
		"finance": "المحاسب",
		"auditor": "المراجع",
		"hr":      "الموارد البشرية",
	}[role]
	if roleArabic == "" {
		roleArabic = role
	}

	subject := "مرحباً بك في نظام CSO ERP"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #1a56db 0%%, #2563eb 100%%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .header h1 { margin: 0; font-size: 28px; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .welcome-text { font-size: 18px; color: #1f2937; margin-bottom: 20px; }
        .info-box { background: #f0f9ff; border: 1px solid #bae6fd; border-radius: 8px; padding: 20px; margin: 20px 0; }
        .info-row { display: flex; justify-content: space-between; padding: 10px 0; border-bottom: 1px solid #e2e8f0; }
        .info-row:last-child { border-bottom: none; }
        .info-label { color: #6b7280; font-weight: 500; }
        .info-value { color: #1f2937; font-weight: 700; }
        .note { color: #6b7280; font-size: 14px; margin-top: 20px; padding: 15px; background: #fef9c3; border-radius: 6px; border-right: 4px solid #facc15; }
        .footer { background: #374151; color: #9ca3af; padding: 20px; text-align: center; font-size: 12px; border-radius: 0 0 8px 8px; }
        .footer p { margin: 5px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 مرحباً بك في الفريق!</h1>
        </div>
        <div class="content">
            <p class="welcome-text">مرحباً <strong>%s</strong>،</p>
            <p>تم إنشاء حسابك بنجاح في نظام إدارة الموارد المالية (CSO ERP). يسعدنا انضمامك إلينا!</p>
            
            <div class="info-box">
                <h3 style="margin-top: 0; color: #1a56db;">بيانات حسابك:</h3>
                <div class="info-row">
                    <span class="info-label">اسم المستخدم:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">البريد الإلكتروني:</span>
                    <span class="info-value">%s</span>
                </div>
                <div class="info-row">
                    <span class="info-label">الصلاحية:</span>
                    <span class="info-value">%s</span>
                </div>
            </div>
            
            <div class="note">
                💡 <strong>ملاحظة هامة:</strong> للحفاظ على أمان حسابك، يرجى تغيير كلمة المرور عند أول تسجيل دخول وتفعيل خاصية التحقق بخطوتين (2FA).
            </div>
            
            <p style="margin-top: 25px; font-size: 14px; color: #6b7280;">
                تم إرسال هذه الرسالة في: %s
            </p>
        </div>
        <div class="footer">
            <p>هذا بريد آلي، يرجى عدم الرد عليه.</p>
            <p>© منظمة LCO - جميع الحقوق محفوظة</p>
        </div>
    </div>
</body>
</html>`, fullName, username, toEmail, roleArabic, time.Now().Format("2006-01-02 15:04:05"))

	return s.SendEmail(ctx, toEmail, subject, body, true)
}

// SendVerificationCode sends a domain verification code for new organization setup
func (s *EmailService) SendVerificationCode(toEmail, orgName, code string) error {
	if err := validateEmail(toEmail); err != nil {
		return err
	}

	orgName = sanitizeHeader(orgName)

	subject := fmt.Sprintf("رمز التحقق - %s", orgName)
	body := fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #059669 0%%, #10b981 100%%); color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .code-box { background: #059669; color: white; font-size: 36px; letter-spacing: 10px; padding: 25px; text-align: center; margin: 25px 0; border-radius: 10px; font-family: monospace; font-weight: bold; }
        .org-name { background: #f0fdf4; border: 1px solid #86efac; padding: 15px; border-radius: 8px; margin: 15px 0; }
        .warning { color: #dc2626; font-size: 14px; margin-top: 20px; padding: 15px; background: #fef2f2; border-radius: 6px; border-right: 4px solid #dc2626; }
        .footer { background: #374151; color: #9ca3af; padding: 15px; text-align: center; font-size: 12px; border-radius: 0 0 8px 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🏢 تفعيل المنظمة الجديدة</h1>
        </div>
        <div class="content">
            <p>مرحباً،</p>
            <p>تم طلب تفعيل منظمة جديدة في نظام ERP:</p>
            
            <div class="org-name">
                <strong>اسم المنظمة:</strong> %s
            </div>
            
            <p>استخدم رمز التحقق التالي لإكمال عملية التفعيل:</p>
            <div class="code-box">%s</div>
            
            <p>⏱️ <strong>صلاحية هذا الرمز: 5 دقائق فقط</strong></p>
            
            <div class="warning">
                ⚠️ <strong>تحذير:</strong> إذا لم تطلب إنشاء هذه المنظمة، يرجى تجاهل هذا البريد أو التواصل مع الدعم الفني.
            </div>
        </div>
        <div class="footer">
            <p>هذا بريد آلي من نظام ERP</p>
            <p>© LCO - جميع الحقوق محفوظة</p>
        </div>
    </div>
</body>
</html>`, orgName, code)

	ctx := context.Background()
	return s.SendEmail(ctx, toEmail, subject, body, true)
}

// SendTenantWelcomeEmail sends welcome email to new tenant admin with credentials
func (s *EmailService) SendTenantWelcomeEmail(ctx context.Context, toEmail, fullName, orgName, username, password, loginURL string) error {
	if err := validateEmail(toEmail); err != nil {
		return err
	}

	fullName = sanitizeHeader(fullName)
	orgName = sanitizeHeader(orgName)

	subject := fmt.Sprintf("مرحباً بك في نظام %s", orgName)
	body := fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #1a56db 0%%, #2563eb 100%%); color: white; padding: 30px; text-align: center; border-radius: 8px 8px 0 0; }
        .header h1 { margin: 0; font-size: 28px; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; }
        .welcome-text { font-size: 18px; color: #1f2937; margin-bottom: 20px; }
        .credentials-box { background: #f0f9ff; border: 1px solid #bae6fd; border-radius: 10px; padding: 25px; margin: 20px 0; }
        .credential-row { display: flex; justify-content: space-between; padding: 12px 0; border-bottom: 1px solid #e2e8f0; }
        .credential-row:last-child { border-bottom: none; }
        .credential-label { color: #6b7280; font-weight: 500; }
        .credential-value { color: #1f2937; font-weight: 700; font-family: monospace; background: #e0f2fe; padding: 4px 10px; border-radius: 4px; }
        .login-btn { display: inline-block; background: #1a56db; color: white; padding: 15px 30px; text-decoration: none; border-radius: 8px; font-weight: bold; margin: 20px 0; }
        .warning { color: #b45309; font-size: 14px; margin-top: 20px; padding: 15px; background: #fef3c7; border-radius: 6px; border-right: 4px solid #f59e0b; }
        .footer { background: #374151; color: #9ca3af; padding: 20px; text-align: center; font-size: 12px; border-radius: 0 0 8px 8px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎉 مرحباً بك في %s!</h1>
        </div>
        <div class="content">
            <p class="welcome-text">مرحباً <strong>%s</strong>،</p>
            <p>تم إنشاء حسابك كمدير نظام للمنظمة. فيما يلي بيانات الدخول الخاصة بك:</p>
            
            <div class="credentials-box">
                <h3 style="margin-top: 0; color: #1a56db;">🔑 بيانات الدخول</h3>
                <div class="credential-row">
                    <span class="credential-label">اسم المستخدم:</span>
                    <span class="credential-value">%s</span>
                </div>
                <div class="credential-row">
                    <span class="credential-label">كلمة المرور:</span>
                    <span class="credential-value">%s</span>
                </div>
                <div class="credential-row">
                    <span class="credential-label">رابط الدخول:</span>
                    <span class="credential-value">%s</span>
                </div>
            </div>
            
            <center>
                <a href="%s" class="login-btn">🚀 الدخول إلى النظام</a>
            </center>
            
            <div class="warning">
                ⚠️ <strong>هام جداً:</strong> يرجى تغيير كلمة المرور فور تسجيل الدخول الأول، وتفعيل خاصية التحقق بخطوتين (2FA) لحماية حسابك.
            </div>
            
            <p style="margin-top: 25px; font-size: 13px; color: #6b7280;">
                تم إرسال هذه الرسالة في: %s
            </p>
        </div>
        <div class="footer">
            <p>هذا بريد آلي، يرجى عدم الرد عليه.</p>
            <p>© %s - جميع الحقوق محفوظة</p>
        </div>
    </div>
</body>
</html>`, orgName, fullName, username, password, loginURL, loginURL, time.Now().Format("2006-01-02 15:04:05"), orgName)

	return s.SendEmail(ctx, toEmail, subject, body, true)
}

// SendBulkEmail sends personalized emails to multiple recipients
func (s *EmailService) SendBulkEmail(ctx context.Context, recipients []EmailRecipient, subject, message string) *BulkEmailResult {
	result := &BulkEmailResult{
		Total:     len(recipients),
		Sent:      0,
		Failed:    0,
		FailedIDs: []string{},
	}

	for _, recipient := range recipients {
		if err := validateEmail(recipient.Email); err != nil {
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, recipient.ID)
			continue
		}

		body := s.buildBulkEmailBody(recipient.Name, message)
		err := s.SendEmail(ctx, recipient.Email, subject, body, true)
		if err != nil {
			log.Printf("Failed to send email to %s: %v", recipient.Email, err)
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, recipient.ID)
		} else {
			result.Sent++
		}
	}

	return result
}

// buildBulkEmailBody creates a beautiful HTML email body for bulk messages
func (s *EmailService) buildBulkEmailBody(recipientName, message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; margin: 0; padding: 0; background-color: #f4f4f4; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #1a56db 0%%, #3b82f6 100%%); color: white; padding: 25px 30px; text-align: center; border-radius: 12px 12px 0 0; }
        .header h1 { margin: 0; font-size: 24px; font-weight: 600; }
        .header .subtitle { margin-top: 8px; font-size: 14px; opacity: 0.9; }
        .content { background: #ffffff; padding: 30px; border: 1px solid #e5e7eb; border-top: none; }
        .greeting { font-size: 18px; color: #1f2937; margin-bottom: 20px; }
        .message-box { background: linear-gradient(135deg, #f0f9ff 0%%, #e0f2fe 100%%); border: 1px solid #bae6fd; border-radius: 10px; padding: 25px; margin: 20px 0; line-height: 1.8; color: #0c4a6e; font-size: 15px; }
        .divider { height: 1px; background: linear-gradient(to right, transparent, #e5e7eb, transparent); margin: 25px 0; }
        .footer { background: #1f2937; color: #9ca3af; padding: 20px; text-align: center; font-size: 12px; border-radius: 0 0 12px 12px; }
        .footer p { margin: 5px 0; }
        .logo { font-size: 28px; margin-bottom: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">📧</div>
            <h1>رسالة من منظمة LCO</h1>
            <div class="subtitle">نظام إدارة الموارد المالية</div>
        </div>
        <div class="content">
            <p class="greeting">مرحباً <strong>%s</strong>،</p>
            
            <div class="message-box">
                %s
            </div>
            
            <div class="divider"></div>
            
            <p style="font-size: 13px; color: #6b7280;">
                تم إرسال هذه الرسالة في: %s
            </p>
        </div>
        <div class="footer">
            <p>هذا بريد آلي من نظام CSO ERP</p>
            <p>© منظمة LCO - جميع الحقوق محفوظة</p>
        </div>
    </div>
</body>
</html>`, recipientName, message, time.Now().Format("2006-01-02 15:04:05"))
}

type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unknown from server: %s", string(fromServer))
		}
	}
	return nil, nil
}

func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string, isHTML bool) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if err := validateEmail(to); err != nil {
		return err
	}

	if len(body) > MaxEmailSize {
		return fmt.Errorf("email body too large")
	}

	if s.mock {
		log.Printf("[MOCK EMAIL] To: %s | Subject: %s", to, subject)
		return nil
	}

	contentType := "text/plain; charset=UTF-8"
	if isHTML {
		contentType = "text/html; charset=UTF-8"
	}

	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	var message strings.Builder
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", sanitizeHeader(s.config.FromName), s.config.FromAddress))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
	message.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	message.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	message.WriteString(fmt.Sprintf("Message-ID: <%d.%s@%s>\r\n", time.Now().UnixNano(), "cso-erp", "lco.org.ly"))
	message.WriteString("\r\n")
	message.WriteString(body)

	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	log.Printf("[SMTP DEBUG] Connecting to %s...", addr)

	var conn net.Conn
	var client *smtp.Client

	// Port 465 uses Implicit TLS (direct SSL connection)
	// Port 587 uses STARTTLS (upgrade after connect)
	if s.config.SMTPPort == "465" && s.config.UseTLS {
		log.Printf("[SMTP DEBUG] Using Implicit TLS (Port 465)...")
		tlsConfig := &tls.Config{
			ServerName:         s.config.SMTPHost,
			InsecureSkipVerify: s.config.SkipVerify,
			MinVersion:         MinTLSVersion,
		}
		dialer := &tls.Dialer{
			NetDialer: &net.Dialer{Timeout: SMTPDialTimeout},
			Config:    tlsConfig,
		}
		var err error
		conn, err = dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			log.Printf("[SMTP ERROR] TLS Connection failed: %v", err)
			return err
		}
		log.Printf("[SMTP DEBUG] TLS Connected successfully")

		client, err = smtp.NewClient(conn, s.config.SMTPHost)
		if err != nil {
			conn.Close()
			log.Printf("[SMTP ERROR] Client creation failed: %v", err)
			return err
		}
	} else {
		// Standard connection for port 587 with STARTTLS
		dialer := &net.Dialer{Timeout: SMTPDialTimeout}
		var err error
		conn, err = dialer.DialContext(ctx, "tcp", addr)
		if err != nil {
			log.Printf("[SMTP ERROR] Connection failed: %v", err)
			return err
		}
		log.Printf("[SMTP DEBUG] Connected successfully")

		client, err = smtp.NewClient(conn, s.config.SMTPHost)
		if err != nil {
			conn.Close()
			log.Printf("[SMTP ERROR] Client creation failed: %v", err)
			return err
		}
	}
	defer client.Close()

	var err error
	if err = client.Hello("lco.org.ly"); err != nil {
		log.Printf("[SMTP ERROR] HELLO failed: %v", err)
		return err
	}
	log.Printf("[SMTP DEBUG] HELLO OK")

	// Skip STARTTLS for port 465 (already using implicit TLS)
	// Only use STARTTLS for port 587 or other non-SSL ports
	if s.config.UseTLS && s.config.SMTPPort != "465" {
		log.Printf("[SMTP DEBUG] Starting TLS (STARTTLS)...")
		tlsConfig := &tls.Config{
			ServerName:         s.config.SMTPHost,
			InsecureSkipVerify: s.config.SkipVerify,
			MinVersion:         MinTLSVersion,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			log.Printf("[SMTP ERROR] STARTTLS failed: %v", err)
			return err
		}
		log.Printf("[SMTP DEBUG] TLS OK")
	} else if s.config.SMTPPort == "465" {
		log.Printf("[SMTP DEBUG] Using Implicit TLS (Port 465) - STARTTLS skipped")
	}

	if s.config.Username != "" {
		log.Printf("[SMTP DEBUG] Authenticating as %s...", s.config.Username)
		auth := LoginAuth(s.config.Username, s.config.Password)
		if err = client.Auth(auth); err != nil {
			log.Printf("[SMTP DEBUG] LOGIN auth failed, trying PLAIN: %v", err)
			auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)
			if err = client.Auth(auth); err != nil {
				log.Printf("[SMTP ERROR] Authentication failed: %v", err)
				return err
			}
		}
		log.Printf("[SMTP DEBUG] Auth OK")
	}

	log.Printf("[SMTP DEBUG] MAIL FROM: %s", s.config.FromAddress)
	if err = client.Mail(s.config.FromAddress); err != nil {
		log.Printf("[SMTP ERROR] MAIL FROM failed: %v", err)
		return err
	}

	log.Printf("[SMTP DEBUG] RCPT TO: %s", to)
	if err = client.Rcpt(to); err != nil {
		log.Printf("[SMTP ERROR] RCPT TO failed: %v", err)
		return err
	}

	log.Printf("[SMTP DEBUG] Sending DATA...")
	w, err := client.Data()
	if err != nil {
		log.Printf("[SMTP ERROR] DATA command failed: %v", err)
		return err
	}
	_, err = w.Write([]byte(message.String()))
	if err != nil {
		log.Printf("[SMTP ERROR] Write failed: %v", err)
		return err
	}
	err = w.Close()
	if err != nil {
		log.Printf("[SMTP ERROR] Close failed: %v", err)
		return err
	}
	log.Printf("[SMTP DEBUG] DATA OK")

	log.Printf("[SMTP SUCCESS] Email sent to %s", to)
	return client.Quit()
}

// ============================================================================
// HIERARCHICAL SMTP SUPPORT (Tenant-Specific Configuration)
// ============================================================================

// TenantEmailConfig represents SMTP config from tenant database
type TenantEmailConfig struct {
	SMTPHost    string
	SMTPPort    int
	Username    string
	Password    string // Decrypted password
	FromAddress string
	FromName    string
	UseTLS      bool
}

// NewEmailServiceWithTenantConfig creates EmailService with tenant-specific SMTP
func NewEmailServiceWithTenantConfig(tenantConfig *TenantEmailConfig) *EmailService {
	if tenantConfig == nil || tenantConfig.SMTPHost == "" {
		log.Println("Tenant SMTP not configured, will use master SMTP")
		return nil // Signal to use master
	}

	config := &EmailConfig{
		SMTPHost:    tenantConfig.SMTPHost,
		SMTPPort:    fmt.Sprintf("%d", tenantConfig.SMTPPort),
		Username:    tenantConfig.Username,
		Password:    tenantConfig.Password,
		FromAddress: tenantConfig.FromAddress,
		FromName:    tenantConfig.FromName,
		UseTLS:      tenantConfig.UseTLS,
		SkipVerify:  false, // Always verify in production
	}

	if !emailRegex.MatchString(config.FromAddress) {
		log.Printf("Invalid tenant FROM address: %s", config.FromAddress)
		return nil
	}

	return &EmailService{config: config, mock: false}
}

// GetHierarchicalEmailService returns tenant EmailService or falls back to master
// tenantService = EmailService from tenant SMTP settings (nil if not configured)
// masterService = EmailService from master/system SMTP settings
func GetHierarchicalEmailService(tenantService, masterService *EmailService) *EmailService {
	if tenantService != nil && !tenantService.mock {
		log.Println("Using tenant-specific SMTP configuration")
		return tenantService
	}
	log.Println("Using master SMTP configuration")
	return masterService
}

// SendWithTemplate sends email using a custom template with variable substitution
func (s *EmailService) SendWithTemplate(ctx context.Context, toEmail, subject, bodyHTML string, variables map[string]string) error {
	if err := validateEmail(toEmail); err != nil {
		return err
	}

	// Replace template variables
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		subject = strings.ReplaceAll(subject, placeholder, sanitizeHeader(value))
		bodyHTML = strings.ReplaceAll(bodyHTML, placeholder, value)
	}

	return s.SendEmail(ctx, toEmail, subject, bodyHTML, true)
}

// SendBulkWithTemplate sends bulk emails with template and per-recipient variables
func (s *EmailService) SendBulkWithTemplate(ctx context.Context, recipients []EmailRecipient, subject, bodyHTML string, commonVars map[string]string) *BulkEmailResult {
	result := &BulkEmailResult{
		Total:     len(recipients),
		FailedIDs: []string{},
	}

	for _, recipient := range recipients {
		// Merge common vars with recipient-specific vars
		vars := make(map[string]string)
		for k, v := range commonVars {
			vars[k] = v
		}
		vars["user_name"] = recipient.Name
		vars["email"] = recipient.Email

		err := s.SendWithTemplate(ctx, recipient.Email, subject, bodyHTML, vars)
		if err != nil {
			log.Printf("Failed to send to %s: %v", recipient.Email, err)
			result.Failed++
			result.FailedIDs = append(result.FailedIDs, recipient.ID)
		} else {
			result.Sent++
		}

		// Rate limiting
		time.Sleep(time.Second / MaxEmailsPerMinute * 60)
	}

	return result
}

// SendDeletionOTPEmail sends OTP for organization deletion confirmation
func (s *EmailService) SendDeletionOTPEmail(ctx context.Context, toEmail, adminName, orgName, code string) error {
	if err := validateEmail(toEmail); err != nil {
		return err
	}

	adminName = sanitizeHeader(adminName)
	orgName = sanitizeHeader(orgName)

	subject := "⚠️ تأكيد حذف منظمة - رمز التحقق"
	body := fmt.Sprintf(`<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; margin: 0; padding: 0; background-color: #1a1a1a; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #dc2626 0%%, #b91c1c 100%%); color: white; padding: 30px; text-align: center; border-radius: 12px 12px 0 0; }
        .header h1 { margin: 0; font-size: 24px; }
        .header .badge { display: inline-block; background: rgba(255,255,255,0.2); padding: 5px 15px; border-radius: 20px; font-size: 12px; margin-top: 10px; }
        .content { background: #292929; padding: 30px; border: 1px solid #404040; color: #e2e8f0; }
        .danger-box { background: rgba(220, 38, 38, 0.2); border: 2px solid #dc2626; border-radius: 10px; padding: 20px; margin: 20px 0; }
        .danger-box h3 { color: #fca5a5; margin-top: 0; }
        .org-name { background: #3f3f3f; padding: 15px; border-radius: 8px; margin: 15px 0; text-align: center; font-size: 20px; color: #fbbf24; }
        .code-box { background: linear-gradient(135deg, #dc2626 0%%, #b91c1c 100%%); color: white; font-size: 36px; letter-spacing: 10px; padding: 25px; text-align: center; margin: 25px 0; border-radius: 12px; font-family: 'Courier New', monospace; font-weight: bold; }
        .warning-list { background: rgba(251, 191, 36, 0.1); border: 1px solid #fbbf24; border-radius: 8px; padding: 15px 25px; color: #fcd34d; margin: 20px 0; }
        .warning-list li { margin: 10px 0; }
        .footer { background: #0f0f0f; color: #6b7280; padding: 20px; text-align: center; font-size: 12px; border-radius: 0 0 12px 12px; }
        .shield { font-size: 48px; margin-bottom: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="shield">🗑️</div>
            <h1>تأكيد حذف منظمة</h1>
            <div class="badge">⚠️ عملية لا يمكن التراجع عنها</div>
        </div>
        <div class="content">
            <p>مرحباً <strong>%s</strong>،</p>
            <p>تم طلب حذف المنظمة التالية نهائياً:</p>
            
            <div class="org-name">
                🏢 %s
            </div>
            
            <div class="danger-box">
                <h3>⚠️ تحذير شديد الأهمية</h3>
                <p>هذه العملية ستؤدي إلى:</p>
                <ul>
                    <li>حذف قاعدة بيانات المنظمة بالكامل</li>
                    <li>حذف جميع بيانات المستخدمين والمعاملات</li>
                    <li>حذف جميع الملفات والمرفقات</li>
                    <li><strong>لا يمكن استرجاع أي بيانات بعد الحذف!</strong></li>
                </ul>
            </div>
            
            <p style="text-align: center;">إذا كنت متأكداً، استخدم رمز التحقق التالي:</p>
            <div class="code-box">%s</div>
            <p style="text-align: center; color: #94a3b8;">⏱️ صلاحية الرمز: <strong style="color: #ef4444;">10 دقائق</strong></p>
            
            <div class="warning-list">
                <strong>🔒 إجراءات الأمان:</strong>
                <ul>
                    <li>لا تشارك هذا الرمز مع أي شخص</li>
                    <li>إذا لم تطلب هذه العملية، تواصل مع الدعم فوراً</li>
                    <li>سيتم تسجيل هذه العملية في سجل المراجعة</li>
                </ul>
            </div>
            
            <p style="font-size: 13px; color: #6b7280; text-align: center;">
                تم إرسال هذا الطلب في: %s
            </p>
        </div>
        <div class="footer">
            <p>هذا بريد آلي عالي الأمان من نظام CSO ERP</p>
            <p>© منظمة LCO - جميع الحقوق محفوظة</p>
        </div>
    </div>
</body>
</html>`, adminName, orgName, code, time.Now().Format("2006-01-02 15:04:05"))

	return s.SendEmail(ctx, toEmail, subject, body, true)
}

// GetConfig returns the email configuration (useful for testing)
func (s *EmailService) GetConfig() *EmailConfig {
	return s.config
}

// IsMock returns whether the service is in mock mode
func (s *EmailService) IsMock() bool {
	return s.mock
}
