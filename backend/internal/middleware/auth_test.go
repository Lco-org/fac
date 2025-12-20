package middleware

import (
	"cso-erp/internal/config"
	"cso-erp/internal/models"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Test JWT configuration
var testJWTConfig = &config.JWTConfig{
	Secret:        "test-secret-key-must-be-32-chars-long!!",
	Expiry:        time.Hour,
	RefreshExpiry: time.Hour * 24,
}

// createTestToken creates a valid JWT token for testing
func createTestToken(user *models.User, cfg *config.JWTConfig, expired bool) string {
	var expTime time.Time
	if expired {
		expTime = time.Now().Add(-time.Hour) // Expired 1 hour ago
	} else {
		expTime = time.Now().Add(time.Hour) // Valid for 1 hour
	}

	claims := JWTClaims{
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "cso-erp",
			Subject:   user.ID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(cfg.Secret))
	return tokenString
}

// ===============================================
// AUTH MIDDLEWARE TESTS
// ===============================================

func TestAuthMiddleware_NoToken(t *testing.T) {
	app := fiber.New()

	app.Use(AuthMiddleware(testJWTConfig))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	app := fiber.New()

	app.Use(AuthMiddleware(testJWTConfig))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	app := fiber.New()

	app.Use(AuthMiddleware(testJWTConfig))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Role:         models.RoleAdmin,
		TokenVersion: 1,
	}
	expiredToken := createTestToken(user, testJWTConfig, true)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_ValidToken_Header(t *testing.T) {
	app := fiber.New()

	var capturedUserID uuid.UUID
	var capturedRole models.UserRole

	app.Use(AuthMiddleware(testJWTConfig))
	app.Get("/protected", func(c *fiber.Ctx) error {
		capturedUserID = GetUserID(c)
		capturedRole = GetUserRole(c)
		return c.SendString("OK")
	})

	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Role:         models.RoleAdmin,
		TokenVersion: 1,
	}
	validToken := createTestToken(user, testJWTConfig, false)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if capturedUserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, capturedUserID)
	}

	if capturedRole != models.RoleAdmin {
		t.Errorf("Expected role admin, got %s", capturedRole)
	}
}

func TestAuthMiddleware_ValidToken_Cookie(t *testing.T) {
	app := fiber.New()

	app.Use(AuthMiddleware(testJWTConfig))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Role:         models.RoleFinance,
		TokenVersion: 1,
	}
	validToken := createTestToken(user, testJWTConfig, false)

	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Cookie", "auth_token="+validToken)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestAuthMiddleware_SkipsSystemAuthRoutes(t *testing.T) {
	app := fiber.New()

	app.Use(AuthMiddleware(testJWTConfig))
	app.Post("/api/v1/system-auth/login", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("POST", "/api/v1/system-auth/login", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	// Should be OK without token because system-auth is skipped
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

// ===============================================
// TOKEN GENERATION TESTS
// ===============================================

func TestGenerateToken_Success(t *testing.T) {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Role:         models.RoleAdmin,
		TokenVersion: 1,
	}

	token, err := GenerateToken(user, testJWTConfig)

	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	// Verify token can be parsed
	parsedToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(testJWTConfig.Secret), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		t.Error("Generated token should be valid")
	}

	claims, ok := parsedToken.Claims.(*JWTClaims)
	if !ok {
		t.Fatal("Failed to get claims")
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.Role != models.RoleAdmin {
		t.Errorf("Expected role admin, got %s", claims.Role)
	}
}

func TestGenerateTokenWithTenant_Success(t *testing.T) {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Role:         models.RoleFinance,
		TokenVersion: 2,
	}
	tenantDB := "org_acme_db"

	token, err := GenerateTokenWithTenant(user, testJWTConfig, tenantDB)

	if err != nil {
		t.Fatalf("GenerateTokenWithTenant failed: %v", err)
	}

	// Verify tenant DB is in claims
	parsedToken, _ := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(testJWTConfig.Secret), nil
	})

	claims := parsedToken.Claims.(*JWTClaims)

	if claims.TenantDB != tenantDB {
		t.Errorf("Expected tenant DB %s, got %s", tenantDB, claims.TenantDB)
	}
}

// ===============================================
// HELPER FUNCTION TESTS
// ===============================================

func TestGetClientIP_XForwardedFor(t *testing.T) {
	app := fiber.New()

	var capturedIP string

	app.Get("/test", func(c *fiber.Ctx) error {
		capturedIP = GetClientIP(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.100, 10.0.0.1")
	app.Test(req)

	if capturedIP != "192.168.1.100" {
		t.Errorf("Expected IP 192.168.1.100, got %s", capturedIP)
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	app := fiber.New()

	var capturedIP string

	app.Get("/test", func(c *fiber.Ctx) error {
		capturedIP = GetClientIP(c)
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Real-IP", "10.0.0.50")
	app.Test(req)

	if capturedIP != "10.0.0.50" {
		t.Errorf("Expected IP 10.0.0.50, got %s", capturedIP)
	}
}

func TestGetUserAgent(t *testing.T) {
	app := fiber.New()

	var capturedUA string

	app.Get("/test", func(c *fiber.Ctx) error {
		capturedUA = GetUserAgent(c)
		return c.SendString("OK")
	})

	expectedUA := "Mozilla/5.0 (Test)"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", expectedUA)
	app.Test(req)

	if capturedUA != expectedUA {
		t.Errorf("Expected UA %s, got %s", expectedUA, capturedUA)
	}
}

// ===============================================
// BENCHMARK TESTS
// ===============================================

func BenchmarkAuthMiddleware(b *testing.B) {
	app := fiber.New()

	app.Use(AuthMiddleware(testJWTConfig))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	user := &models.User{
		ID:           uuid.New(),
		Username:     "benchuser",
		Role:         models.RoleAdmin,
		TokenVersion: 1,
	}
	token := createTestToken(user, testJWTConfig, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		app.Test(req)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	user := &models.User{
		ID:           uuid.New(),
		Username:     "benchuser",
		Role:         models.RoleAdmin,
		TokenVersion: 1,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateToken(user, testJWTConfig)
	}
}
