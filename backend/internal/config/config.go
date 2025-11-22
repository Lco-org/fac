package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Security SecurityConfig
	CORS     CORSConfig
	Upload   UploadConfig
}

type AppConfig struct {
	Env  string
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
}

type SecurityConfig struct {
	BcryptCost      int
	RateLimitMax    int
	RateLimitWindow time.Duration
}

type CORSConfig struct {
	Origin      string
	Credentials bool
}

type UploadConfig struct {
	MaxFileSize int64
	UploadDir   string
}

func Load() *Config {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	jwtExpiry, _ := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	jwtRefreshExpiry, _ := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	rateLimitWindow, _ := time.ParseDuration(getEnv("RATE_LIMIT_WINDOW", "1m"))

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
			BcryptCost:      getEnvAsInt("BCRYPT_COST", 10),
			RateLimitMax:    getEnvAsInt("RATE_LIMIT_MAX", 100),
			RateLimitWindow: rateLimitWindow,
		},
		CORS: CORSConfig{
			Origin:      getEnv("CORS_ORIGIN", "http://localhost:5173"),
			Credentials: getEnvAsBool("CORS_CREDENTIALS", true),
		},
		Upload: UploadConfig{
			MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE", 10485760),
			UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
		},
	}
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
