package config

import (
	"encoding/hex"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Security SecurityConfig
	CORS     CORSConfig
	Upload   UploadConfig
	Email    EmailConfig
}

// AppConfig holds application settings
type AppConfig struct {
	Env  string
	Port string
	Host string
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// JWTConfig holds JWT token settings
type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
	// EncryptionKey is a 32-byte key for AES-256-GCM encryption
	// Must be hex-encoded (64 characters)
	EncryptionKey []byte

	// Argon2 password hashing parameters
	Argon2Memory      uint32 // Memory in KB (default: 65536 = 64MB)
	Argon2Iterations  uint32 // Time cost (default: 3)
	Argon2Parallelism uint8  // CPU threads (default: 4)

	// Rate limiting
	RateLimitMax    int
	RateLimitWindow time.Duration
	CookieSecure    bool
}

// CORSConfig holds CORS settings
type CORSConfig struct {
	Origin      string
	Credentials bool
}

// UploadConfig holds file upload settings
type UploadConfig struct {
	MaxFileSize int64
	UploadDir   string
}

// EmailConfig holds SMTP email settings for 2FA
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

// Load loads configuration from environment variables
func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtExpiry, _ := time.ParseDuration(getEnv("JWT_EXPIRY", "3h"))
	jwtRefreshExpiry, _ := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	rateLimitWindow, _ := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))

	// Parse encryption key from hex
	encryptionKey := parseEncryptionKey(getEnv("ENCRYPTION_KEY", ""))

	return &Config{
		App: AppConfig{
			Env:  getEnv("APP_ENV", "development"),
			Port: getEnv("APP_PORT", "8080"),
			Host: getEnv("APP_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "cso_erp"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "change-this-secret-key"),
			Expiry:        jwtExpiry,
			RefreshExpiry: jwtRefreshExpiry,
		},
		Security: SecurityConfig{
			EncryptionKey:     encryptionKey,
			Argon2Memory:      uint32(getEnvAsInt("ARGON2_MEMORY", 65536)),
			Argon2Iterations:  uint32(getEnvAsInt("ARGON2_ITERATIONS", 3)),
			Argon2Parallelism: uint8(getEnvAsInt("ARGON2_PARALLELISM", 4)),
			RateLimitMax:      getEnvAsInt("RATE_LIMIT_MAX", 100),
			RateLimitWindow:   rateLimitWindow,
			CookieSecure:      getEnvAsBool("COOKIE_SECURE", true),
		},
		CORS: CORSConfig{
			Origin:      getEnv("CORS_ORIGIN", "http://localhost:5173"),
			Credentials: getEnvAsBool("CORS_CREDENTIALS", true),
		},
		Upload: UploadConfig{
			MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE", 10485760),
			UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		},
		Email: EmailConfig{
			SMTPHost:    getEnv("SMTP_HOST", ""),
			SMTPPort:    getEnv("SMTP_PORT", "587"),
			Username:    getEnv("SMTP_USERNAME", ""),
			Password:    getEnv("SMTP_PASSWORD", ""),
			FromAddress: getEnv("SMTP_FROM_ADDRESS", "noreply@lco.org.ly"),
			FromName:    getEnv("SMTP_FROM_NAME", "CSO ERP System"),
			UseTLS:      getEnvAsBool("SMTP_USE_TLS", true),
			SkipVerify:  getEnvAsBool("SMTP_SKIP_VERIFY", false),
		},
	}
}

// parseEncryptionKey parses a hex-encoded encryption key
// Returns a 32-byte key for AES-256 or nil if not configured
func parseEncryptionKey(hexKey string) []byte {
	if hexKey == "" {
		log.Println("⚠️  ENCRYPTION_KEY not set - data encryption disabled")
		return nil
	}

	if len(hexKey) != 64 {
		log.Println("⚠️  ENCRYPTION_KEY must be 64 hex characters (32 bytes)")
		return nil
	}

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		log.Println("⚠️  ENCRYPTION_KEY is not valid hex")
		return nil
	}

	return key
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
