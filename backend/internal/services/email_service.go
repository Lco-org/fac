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

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	SMTPHost    string
	SMTPPort    string
	Username    string
	Password    string
	FromAddress string
	FromName    string
	UseTLS      bool
	SkipVerify  bool // For development only - NEVER use in production
}

// EmailService handles sending emails securely
type EmailService struct {
	config *EmailConfig
	mock   bool // If true, just logs emails instead of sending
}

// Security constants
const (
	// Maximum email size to prevent DoS
	MaxEmailSize = 10 * 1024 * 1024 // 10MB

	// Connection timeout
	SMTPDialTimeout = 30 * time.Second

	// Rate limiting - max emails per minute
	MaxEmailsPerMinute = 10

	// Minimum TLS version
	MinTLSVersion = tls.VersionTLS12
)

// Email validation regex (RFC 5322 simplified)
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmailService creates a new email service
func NewEmailService(config *EmailConfig) *EmailService {
	if config == nil || config.SMTPHost == "" {
		log.Println("⚠️  Email service not configured - running in mock mode (emails will be logged)")
		return &EmailService{config: nil, mock: true}
	}

	// Validate configuration
	if !emailRegex.MatchString(config.FromAddress) {
		log.Printf("⚠️  Invalid FROM email address: %s - using mock mode", config.FromAddress)
		return &EmailService{config: nil, mock: true}
	}

	// Security warning for production
	if config.SkipVerify {
		log.Println("⚠️  WARNING: TLS certificate verification is DISABLED - this is insecure!")
	}

	log.Printf("📧 Email service configured: SMTP=%s:%s, TLS=%v", config.SMTPHost, config.SMTPPort, config.UseTLS)
	return &EmailService{config: config, mock: false}
}

// NewMockEmailService creates a mock email service for testing
func NewMockEmailService() *EmailService {
	return &EmailService{mock: true}
}

// validateEmail validates an email address for security
func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email address is empty")
	}
	if len(email) > 254 {
		return fmt.Errorf("email address too long")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}
	// Prevent header injection attacks
	if strings.ContainsAny(email, "\r\n") {
		return fmt.Errorf("email contains invalid characters")
	}
	return nil
}

// sanitizeHeader sanitizes email header values to prevent header injection
func sanitizeHeader(value string) string {
	// Remove any CR/LF characters to prevent header injection
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	// Limit length
	if len(value) > 998 {
		value = value[:998]
	}
	return value
}

// SendOTPEmail sends a verification code email securely
func (s *EmailService) SendOTPEmail(ctx context.Context, toEmail, toName, code string) error {
	// Validate inputs
	if err := validateEmail(toEmail); err != nil {
		return fmt.Errorf("invalid recipient email: %w", err)
	}

	// Sanitize name to prevent injection
	toName = sanitizeHeader(toName)

	// Validate OTP code format (should be 6 digits only)
	if len(code) != 6 {
		return fmt.Errorf("invalid OTP code length")
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return fmt.Errorf("invalid OTP code format")
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

// LoginAuth implements the LOGIN authentication mechanism
type loginAuth struct {
	username, password string
}

// LoginAuth returns an Auth that implements the LOGIN authentication mechanism.
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

// SendEmail sends an email securely with all protections
func (s *EmailService) SendEmail(ctx context.Context, to, subject, body string, isHTML bool) error {
	// Check context for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Validate recipient
	if err := validateEmail(to); err != nil {
		return fmt.Errorf("invalid recipient: %w", err)
	}

	// Check body size
	if len(body) > MaxEmailSize {
		return fmt.Errorf("email body too large")
	}

	// Mock mode - just log the email
	if s.mock {
		log.Printf(`
╔══════════════════════════════════════════════════════════════════╗
║                    📧 EMAIL (MOCK MODE)                          ║
╠══════════════════════════════════════════════════════════════════╣
║ TO:      %s
║ SUBJECT: %s
╠══════════════════════════════════════════════════════════════════╣
║ BODY: [%d bytes]
╚══════════════════════════════════════════════════════════════════╝
`, to, subject, len(body))
		return nil
	}

	// Build secure email message
	contentType := "text/plain; charset=UTF-8"
	if isHTML {
		contentType = "text/html; charset=UTF-8"
	}

	// RFC 2047 Encode Subject (Base64) for non-ASCII characters
	// Format: =?charset?encoding?encoded_text?=
	encodedSubject := fmt.Sprintf("=?UTF-8?B?%s?=", base64.StdEncoding.EncodeToString([]byte(subject)))

	// Build headers securely
	var message strings.Builder
	message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", sanitizeHeader(s.config.FromName), s.config.FromAddress))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: %s\r\n", contentType))
	message.WriteString("Content-Transfer-Encoding: 8bit\r\n") // Explicitly set encoding
	message.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	message.WriteString(fmt.Sprintf("Message-ID: <%d.%s@%s>\r\n", time.Now().UnixNano(), "cso-erp", "lco.org.ly"))

	// Security headers
	message.WriteString("\r\n")
	message.WriteString(body)

	// Connect to SMTP server with timeout
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)

	dialer := &net.Dialer{Timeout: SMTPDialTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		log.Printf("❌ SMTP connection error: %v", err)
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		conn.Close()
		log.Printf("❌ SMTP client error: %v", err)
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Say hello with proper domain (not localhost)
	// Using lco.org.ly as it matches the sender domain
	if err = client.Hello("lco.org.ly"); err != nil {
		log.Printf("❌ SMTP Hello error: %v", err)
		return err
	}

	// Start TLS with secure configuration
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
			log.Printf("❌ SMTP StartTLS error: %v", err)
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
		log.Println("🔒 TLS connection established")
	}

	// Authenticate securely
	if s.config.Username != "" {
		// Try LOGIN auth first for compatibility
		auth := LoginAuth(s.config.Username, s.config.Password)
		if err = client.Auth(auth); err != nil {
			log.Printf("⚠️ SMTP LOGIN Auth failed: %v. Retrying with PLAIN...", err)
			// Fallback to PLAIN if LOGIN fails
			auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.SMTPHost)
			if err = client.Auth(auth); err != nil {
				log.Printf("❌ SMTP PLAIN Auth error: %v", err)
				return fmt.Errorf("SMTP authentication failed: %w", err)
			}
		}
	}

	// Set sender and recipient
	if err = client.Mail(s.config.FromAddress); err != nil {
		log.Printf("❌ SMTP Mail error: %v", err)
		return err
	}
	if err = client.Rcpt(to); err != nil {
		log.Printf("❌ SMTP Rcpt error: %v", err)
		return err
	}

	// Send message body
	w, err := client.Data()
	if err != nil {
		log.Printf("❌ SMTP Data error: %v", err)
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

	log.Printf("✅ Email sent securely to %s", to)
	return client.Quit()
}
