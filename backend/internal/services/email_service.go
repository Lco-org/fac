package services

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost    string
	SMTPPort    string
	Username    string
	Password    string
	FromAddress string
	FromName    string
	UseTLS      bool
	SkipVerify  bool // For development only
}

// EmailService handles sending emails
type EmailService struct {
	config *EmailConfig
	mock   bool // If true, just logs emails instead of sending
}

// NewEmailService creates a new email service
func NewEmailService(config *EmailConfig) *EmailService {
	if config == nil || config.SMTPHost == "" {
		log.Println("⚠️  Email service not configured - running in mock mode (emails will be logged)")
		return &EmailService{config: nil, mock: true}
	}
	return &EmailService{config: config, mock: false}
}

// NewMockEmailService creates a mock email service for testing
func NewMockEmailService() *EmailService {
	return &EmailService{mock: true}
}

// SendOTPEmail sends a verification code email
func (s *EmailService) SendOTPEmail(ctx context.Context, toEmail, toName, code string) error {
	subject := "رمز التحقق - CSO ERP"

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: 'Segoe UI', Tahoma, Arial, sans-serif; direction: rtl; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1a56db; color: white; padding: 20px; text-align: center; border-radius: 8px 8px 0 0; }
        .content { background: #f9fafb; padding: 30px; border: 1px solid #e5e7eb; }
        .code-box { background: #1a56db; color: white; font-size: 32px; letter-spacing: 8px; padding: 20px; text-align: center; margin: 20px 0; border-radius: 8px; font-family: monospace; }
        .warning { color: #dc2626; font-size: 14px; margin-top: 20px; }
        .footer { background: #374151; color: #9ca3af; padding: 15px; text-align: center; font-size: 12px; border-radius: 0 0 8px 8px; }
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
            <p class="warning">⚠️ إذا لم تطلب تسجيل الدخول، يرجى تجاهل هذا البريد وتغيير كلمة المرور فوراً.</p>
        </div>
        <div class="footer">
            <p>هذا بريد آلي، يرجى عدم الرد عليه.</p>
            <p>© منظمة LCO - جميع الحقوق محفوظة</p>
        </div>
    </div>
</body>
</html>
`, toName, code)

	return s.SendEmail(ctx, toEmail, subject, body, true)
}

// SendEmail sends an email
func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string, isHTML bool) error {
	// Mock mode - just log the email
	if s.mock {
		log.Printf(`
╔══════════════════════════════════════════════════════════════════╗
║                    📧 EMAIL (MOCK MODE)                          ║
╠══════════════════════════════════════════════════════════════════╣
║ TO:      %s
║ SUBJECT: %s
╠══════════════════════════════════════════════════════════════════╣
║ BODY:
%s
╚══════════════════════════════════════════════════════════════════╝
`, to, subject, body)
		return nil
	}

	// Build email message
	contentType := "text/plain"
	if isHTML {
		contentType = "text/html"
	}

	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.FromName, s.config.FromAddress)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = fmt.Sprintf("%s; charset=UTF-8", contentType)

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)

	var auth smtp.Auth
	if s.config.Username != "" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)
	}

	if s.config.UseTLS {
		// TLS connection
		tlsConfig := &tls.Config{
			InsecureSkipVerify: s.config.SkipVerify,
			ServerName:         s.config.SMTPHost,
		}

		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.config.SMTPHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		if auth != nil {
			if err = client.Auth(auth); err != nil {
				return fmt.Errorf("SMTP authentication failed: %w", err)
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

		return client.Quit()
	}

	// Non-TLS connection
	return smtp.SendMail(addr, auth, s.config.FromAddress, []string{to}, []byte(message.String()))
}
