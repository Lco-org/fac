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

	dialer := &net.Dialer{Timeout: SMTPDialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		conn.Close()
		return err
	}
	defer client.Close()

	if err = client.Hello("lco.org.ly"); err != nil {
		return err
	}

	if s.config.UseTLS {
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
			return err
		}
	}

	if s.config.Username != "" {
		auth := LoginAuth(s.config.Username, s.config.Password)
		if err = client.Auth(auth); err != nil {
			log.Printf("LOGIN failed, retrying PLAIN: %v", err)
			auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)
			if err = client.Auth(auth); err != nil {
				return err
			}
		}
	}

	if err = client.Mail(s.config.FromAddress); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(message.String()))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	log.Printf("Email sent to %s", to)
	return client.Quit()
}
