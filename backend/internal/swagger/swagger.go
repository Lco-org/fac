// Package swagger provides Swagger UI and OpenAPI spec handlers for Fiber
package swagger

import (
	"github.com/gofiber/fiber/v2"
)

// Config holds the Swagger configuration
type Config struct {
	// Title of the API
	Title string
	// Version of the API
	Version string
	// Description of the API
	Description string
	// Base path for the API
	BasePath string
	// Host of the API
	Host string
}

// DefaultConfig returns the default Swagger config
func DefaultConfig() Config {
	return Config{
		Title:       "CSO-ERP API",
		Version:     "1.0.0",
		Description: "Multi-Tenant SaaS ERP Platform API",
		BasePath:    "/api/v1",
		Host:        "localhost:8080",
	}
}

// SetupSwaggerRoutes adds Swagger UI and spec routes
func SetupSwaggerRoutes(app *fiber.App, cfg Config) {
	// Serve OpenAPI JSON spec
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.JSON(GenerateSpec(cfg))
	})

	// Serve Swagger UI HTML
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(SwaggerUIHTML(cfg.Title))
	})

	app.Get("/swagger/*", func(c *fiber.Ctx) error {
		return c.Type("html").SendString(SwaggerUIHTML(cfg.Title))
	})
}

// SwaggerUIHTML returns the Swagger UI HTML page
func SwaggerUIHTML(title string) string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + title + ` - API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
    <style>
        body { margin: 0; padding: 0; }
        .swagger-ui .topbar { display: none; }
        .swagger-ui .info .title { font-size: 2rem; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/swagger/doc.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                docExpansion: "list",
                filter: true,
                showRequestHeaders: true,
            });
        };
    </script>
</body>
</html>`
}

// GenerateSpec generates the OpenAPI 3.0 specification
func GenerateSpec(cfg Config) map[string]interface{} {
	return map[string]interface{}{
		"openapi": "3.0.3",
		"info": map[string]interface{}{
			"title":       cfg.Title,
			"version":     cfg.Version,
			"description": cfg.Description,
			"contact": map[string]string{
				"email": "admin@lco.org.ly",
			},
		},
		"servers": []map[string]string{
			{"url": "http://" + cfg.Host + cfg.BasePath, "description": "Development server"},
			{"url": "https://" + cfg.Host + cfg.BasePath, "description": "Production server"},
		},
		"tags": []map[string]string{
			{"name": "Authentication", "description": "User authentication and session management"},
			{"name": "Users", "description": "User management (Admin only)"},
			{"name": "Transactions", "description": "Financial transactions"},
			{"name": "Containers", "description": "Cash and bank account management"},
			{"name": "Projects", "description": "Project management"},
			{"name": "Employees", "description": "Employee management"},
			{"name": "Members", "description": "Membership management"},
			{"name": "Reports", "description": "Financial reports and dashboards"},
			{"name": "HR", "description": "Human resources operations"},
			{"name": "SSA", "description": "Super System Admin operations (Multi-tenant)"},
		},
		"components": map[string]interface{}{
			"securitySchemes": map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
					"description":  "JWT token authentication",
				},
				"CookieAuth": map[string]interface{}{
					"type": "apiKey",
					"in":   "cookie",
					"name": "auth_token",
				},
			},
			"schemas": getSchemas(),
		},
		"paths": getPaths(),
	}
}

func getSchemas() map[string]interface{} {
	return map[string]interface{}{
		"LoginRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"username", "password"},
			"properties": map[string]interface{}{
				"username": map[string]interface{}{
					"type":    "string",
					"example": "admin@lco.org.ly",
				},
				"password": map[string]interface{}{
					"type":    "string",
					"format":  "password",
					"example": "SecurePassword123!",
				},
			},
		},
		"LoginResponse": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"user":  map[string]string{"$ref": "#/components/schemas/User"},
				"token": map[string]interface{}{"type": "string", "example": "eyJhbGciOiJIUzI1NiIs..."},
			},
		},
		"TwoFactorRequired": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"requires_2fa":    map[string]interface{}{"type": "boolean", "example": true},
				"pending_user_id": map[string]interface{}{"type": "string", "format": "uuid"},
				"message":         map[string]interface{}{"type": "string", "example": "Verification code sent to your email"},
			},
		},
		"Verify2FARequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"user_id", "code"},
			"properties": map[string]interface{}{
				"user_id": map[string]interface{}{"type": "string", "format": "uuid"},
				"code":    map[string]interface{}{"type": "string", "example": "123456", "minLength": 6, "maxLength": 6},
			},
		},
		"User": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":                 map[string]interface{}{"type": "string", "format": "uuid"},
				"username":           map[string]interface{}{"type": "string"},
				"email":              map[string]interface{}{"type": "string", "format": "email"},
				"full_name":          map[string]interface{}{"type": "string"},
				"full_name_ar":       map[string]interface{}{"type": "string"},
				"role":               map[string]interface{}{"type": "string", "enum": []string{"admin", "finance", "auditor", "hr"}},
				"two_factor_enabled": map[string]interface{}{"type": "boolean"},
				"is_active":          map[string]interface{}{"type": "boolean"},
				"created_at":         map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"CreateUserRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"username", "email", "password", "full_name", "role"},
			"properties": map[string]interface{}{
				"username":     map[string]interface{}{"type": "string", "minLength": 3},
				"email":        map[string]interface{}{"type": "string", "format": "email"},
				"password":     map[string]interface{}{"type": "string", "format": "password", "minLength": 8},
				"full_name":    map[string]interface{}{"type": "string"},
				"full_name_ar": map[string]interface{}{"type": "string"},
				"role":         map[string]interface{}{"type": "string", "enum": []string{"admin", "finance", "auditor", "hr"}},
			},
		},
		"Transaction": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":                 map[string]interface{}{"type": "string", "format": "uuid"},
				"transaction_number": map[string]interface{}{"type": "string", "example": "TXN-2024-000001"},
				"type":               map[string]interface{}{"type": "string", "enum": []string{"income_grant", "income_donation", "disbursement_custody", "disbursement_salary", "disbursement_operational", "disbursement_donation", "disbursement_bonus", "procurement", "transfer"}},
				"status":             map[string]interface{}{"type": "string", "enum": []string{"pending", "completed", "reversed"}},
				"amount":             map[string]interface{}{"type": "number", "format": "double"},
				"currency":           map[string]interface{}{"type": "string", "enum": []string{"LYD", "USD", "EUR", "GBP"}},
				"exchange_rate":      map[string]interface{}{"type": "number", "format": "double"},
				"description":        map[string]interface{}{"type": "string"},
				"transaction_date":   map[string]interface{}{"type": "string", "format": "date-time"},
			},
		},
		"CreateTransactionRequest": map[string]interface{}{
			"type":     "object",
			"required": []string{"type", "amount", "currency", "container_id", "project_id", "description"},
			"properties": map[string]interface{}{
				"type":             map[string]interface{}{"type": "string", "enum": []string{"income_grant", "income_donation", "disbursement_custody", "disbursement_salary", "disbursement_operational"}},
				"amount":           map[string]interface{}{"type": "number", "format": "double", "minimum": 0.01},
				"currency":         map[string]interface{}{"type": "string", "enum": []string{"LYD", "USD", "EUR", "GBP"}},
				"exchange_rate":    map[string]interface{}{"type": "number", "format": "double", "default": 1.0},
				"container_id":     map[string]interface{}{"type": "string", "format": "uuid"},
				"project_id":       map[string]interface{}{"type": "string", "format": "uuid"},
				"employee_id":      map[string]interface{}{"type": "string", "format": "uuid"},
				"description":      map[string]interface{}{"type": "string"},
				"reference_number": map[string]interface{}{"type": "string"},
			},
		},
		"Container": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":       map[string]interface{}{"type": "string", "format": "uuid"},
				"name":     map[string]interface{}{"type": "string"},
				"type":     map[string]interface{}{"type": "string", "enum": []string{"cash", "bank"}},
				"currency": map[string]interface{}{"type": "string"},
				"balance":  map[string]interface{}{"type": "number", "format": "double"},
			},
		},
		"Project": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":        map[string]interface{}{"type": "string", "format": "uuid"},
				"code":      map[string]interface{}{"type": "string"},
				"name":      map[string]interface{}{"type": "string"},
				"budget":    map[string]interface{}{"type": "number", "format": "double"},
				"currency":  map[string]interface{}{"type": "string"},
				"is_active": map[string]interface{}{"type": "boolean"},
			},
		},
		"Organization": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":                  map[string]interface{}{"type": "string", "format": "uuid"},
				"name":                map[string]interface{}{"type": "string"},
				"slug":                map[string]interface{}{"type": "string"},
				"primary_domain":      map[string]interface{}{"type": "string"},
				"status":              map[string]interface{}{"type": "string", "enum": []string{"pending_verification", "active", "suspended", "deleted"}},
				"subscription_status": map[string]interface{}{"type": "string"},
			},
		},
		"Error": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"error": map[string]interface{}{"type": "string"},
				"code":  map[string]interface{}{"type": "integer"},
			},
		},
	}
}

func getPaths() map[string]interface{} {
	return map[string]interface{}{
		"/auth/login": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Authentication"},
				"summary":     "User login",
				"description": "Authenticate user with email and password. Returns JWT token or requires 2FA verification.",
				"operationId": "login",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{"$ref": "#/components/schemas/LoginRequest"},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Successful login",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]string{"$ref": "#/components/schemas/LoginResponse"},
							},
						},
					},
					"202": map[string]interface{}{
						"description": "2FA verification required",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]string{"$ref": "#/components/schemas/TwoFactorRequired"},
							},
						},
					},
					"401": map[string]interface{}{
						"description": "Invalid credentials",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]string{"$ref": "#/components/schemas/Error"},
							},
						},
					},
				},
			},
		},
		"/auth/verify-2fa": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Authentication"},
				"summary":     "Verify 2FA code",
				"description": "Verify the 6-digit code sent to user's email",
				"operationId": "verify2FA",
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{"$ref": "#/components/schemas/Verify2FARequest"},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Verification successful",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]string{"$ref": "#/components/schemas/LoginResponse"},
							},
						},
					},
					"401": map[string]interface{}{"description": "Invalid or expired code"},
				},
			},
		},
		"/auth/logout": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Authentication"},
				"summary":     "User logout",
				"description": "Invalidate current session",
				"operationId": "logout",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Logged out successfully"},
				},
			},
		},
		"/users": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Users"},
				"summary":     "List all users",
				"description": "Get list of all users (Admin only)",
				"operationId": "listUsers",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of users",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":  "array",
									"items": map[string]string{"$ref": "#/components/schemas/User"},
								},
							},
						},
					},
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"Users"},
				"summary":     "Create new user",
				"description": "Create a new user (Admin only)",
				"operationId": "createUser",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{"$ref": "#/components/schemas/CreateUserRequest"},
						},
					},
				},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{"description": "User created successfully"},
					"400": map[string]interface{}{"description": "Invalid request"},
					"403": map[string]interface{}{"description": "Insufficient permissions"},
				},
			},
		},
		"/transactions": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Transactions"},
				"summary":     "Create transaction",
				"description": "Create a new financial transaction",
				"operationId": "createTransaction",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"requestBody": map[string]interface{}{
					"required": true,
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]string{"$ref": "#/components/schemas/CreateTransactionRequest"},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Transaction created",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]string{"$ref": "#/components/schemas/Transaction"},
							},
						},
					},
				},
			},
		},
		"/transactions/{id}": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Transactions"},
				"summary":     "Get transaction by ID",
				"operationId": "getTransaction",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"parameters": []map[string]interface{}{
					{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string", "format": "uuid"}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "Transaction details",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]string{"$ref": "#/components/schemas/Transaction"},
							},
						},
					},
				},
			},
		},
		"/transactions/{id}/reverse": map[string]interface{}{
			"post": map[string]interface{}{
				"tags":        []string{"Transactions"},
				"summary":     "Reverse transaction",
				"description": "Reverse a completed transaction",
				"operationId": "reverseTransaction",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"parameters": []map[string]interface{}{
					{"name": "id", "in": "path", "required": true, "schema": map[string]string{"type": "string", "format": "uuid"}},
				},
				"requestBody": map[string]interface{}{
					"content": map[string]interface{}{
						"application/json": map[string]interface{}{
							"schema": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"reason": map[string]interface{}{"type": "string", "description": "Reason for reversal"},
								},
							},
						},
					},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Transaction reversed"},
				},
			},
		},
		"/containers": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Containers"},
				"summary":     "List containers",
				"operationId": "listContainers",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of containers",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":  "array",
									"items": map[string]string{"$ref": "#/components/schemas/Container"},
								},
							},
						},
					},
				},
			},
		},
		"/projects": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Projects"},
				"summary":     "List projects",
				"operationId": "listProjects",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of projects",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":  "array",
									"items": map[string]string{"$ref": "#/components/schemas/Project"},
								},
							},
						},
					},
				},
			},
		},
		"/reports/dashboard": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"Reports"},
				"summary":     "Get dashboard summary",
				"description": "Returns financial summary for dashboard",
				"operationId": "getDashboard",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Dashboard data"},
				},
			},
		},
		"/ssa/organizations": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"SSA"},
				"summary":     "List organizations",
				"description": "List all tenant organizations (SSA only)",
				"operationId": "listOrganizations",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"parameters": []map[string]interface{}{
					{"name": "status", "in": "query", "schema": map[string]interface{}{"type": "string", "enum": []string{"pending_verification", "active", "suspended"}}},
					{"name": "limit", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 20}},
					{"name": "offset", "in": "query", "schema": map[string]interface{}{"type": "integer", "default": 0}},
				},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{
						"description": "List of organizations",
						"content": map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type":  "array",
									"items": map[string]string{"$ref": "#/components/schemas/Organization"},
								},
							},
						},
					},
				},
			},
			"post": map[string]interface{}{
				"tags":        []string{"SSA"},
				"summary":     "Create organization",
				"description": "Create a new tenant organization",
				"operationId": "createOrganization",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"responses": map[string]interface{}{
					"201": map[string]interface{}{"description": "Organization created"},
				},
			},
		},
		"/ssa/stats": map[string]interface{}{
			"get": map[string]interface{}{
				"tags":        []string{"SSA"},
				"summary":     "Get organization stats",
				"operationId": "getOrgStats",
				"security":    []map[string][]string{{"BearerAuth": {}}},
				"responses": map[string]interface{}{
					"200": map[string]interface{}{"description": "Organization statistics"},
				},
			},
		},
	}
}
