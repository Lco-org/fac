// Package docs CSO-ERP API Documentation
//
// This is the API documentation for CSO-ERP Multi-Tenant SaaS Platform.
// The API provides endpoints for financial management, HR, projects, and multi-tenant administration.
//
//	Schemes: http, https
//	Host: localhost:8080
//	BasePath: /api/v1
//	Version: 1.0.0
//	License: Proprietary
//	Contact: admin@lco.org.ly
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	SecurityDefinitions:
//	  BearerAuth:
//	    type: apiKey
//	    name: Authorization
//	    in: header
//	    description: "JWT token with Bearer prefix. Example: 'Bearer eyJhbGciOiJIUzI1NiIs...'"
//	  CookieAuth:
//	    type: apiKey
//	    name: auth_token
//	    in: cookie
//	    description: "JWT token stored in cookie"
//
// swagger:meta
package docs

// ============================================================================
// COMMON RESPONSE TYPES
// ============================================================================

// Generic error response
// swagger:response errorResponse
type ErrorResponse struct {
	// in: body
	Body struct {
		// Error message
		// example: Invalid request parameters
		Error string `json:"error"`
		// HTTP status code
		// example: 400
		Code int `json:"code,omitempty"`
	}
}

// Generic success message response
// swagger:response messageResponse
type MessageResponse struct {
	// in: body
	Body struct {
		// Success message
		// example: Operation completed successfully
		Message string `json:"message"`
	}
}

// ============================================================================
// AUTH ENDPOINTS
// ============================================================================

// Login request
// swagger:parameters loginRequest
type LoginRequestParams struct {
	// Login credentials
	// in: body
	// required: true
	Body struct {
		// Username or email
		// required: true
		// example: admin@lco.org.ly
		Username string `json:"username"`
		// Password
		// required: true
		// example: SecurePassword123!
		Password string `json:"password"`
	}
}

// Login response with user and token
// swagger:response loginResponse
type LoginResponse struct {
	// in: body
	Body struct {
		// User information
		User UserInfo `json:"user"`
		// JWT token
		// example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
		Token string `json:"token"`
	}
}

// 2FA required response
// swagger:response twoFactorResponse
type TwoFactorResponse struct {
	// in: body
	Body struct {
		// Indicates 2FA is required
		// example: true
		Requires2FA bool `json:"requires_2fa"`
		// Pending user ID for 2FA verification
		// example: 550e8400-e29b-41d4-a716-446655440000
		PendingUserID string `json:"pending_user_id"`
		// Message
		// example: Verification code sent to your email
		Message string `json:"message"`
	}
}

// 2FA verification request
// swagger:parameters verify2FARequest
type Verify2FARequestParams struct {
	// 2FA verification data
	// in: body
	// required: true
	Body struct {
		// User ID from login response
		// required: true
		// example: 550e8400-e29b-41d4-a716-446655440000
		UserID string `json:"user_id"`
		// 6-digit verification code
		// required: true
		// example: 123456
		Code string `json:"code"`
	}
}

// User information
type UserInfo struct {
	// User ID
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// Username
	// example: admin
	Username string `json:"username"`
	// Email address
	// example: admin@lco.org.ly
	Email string `json:"email"`
	// Full name
	// example: Ahmed Mohamed
	FullName string `json:"full_name"`
	// Full name in Arabic
	// example: أحمد محمد
	FullNameAr string `json:"full_name_ar,omitempty"`
	// User role
	// example: admin
	// enum: admin,finance,auditor,hr
	Role string `json:"role"`
	// 2FA enabled
	// example: true
	TwoFactorEnabled bool `json:"two_factor_enabled"`
	// Account active status
	// example: true
	IsActive bool `json:"is_active"`
}

// ============================================================================
// TRANSACTION ENDPOINTS
// ============================================================================

// Create transaction request
// swagger:parameters createTransactionRequest
type CreateTransactionRequestParams struct {
	// Transaction data
	// in: body
	// required: true
	Body struct {
		// Transaction type
		// required: true
		// example: income_grant
		// enum: income_grant,income_donation,disbursement_custody,disbursement_salary,disbursement_operational,disbursement_donation,disbursement_bonus,procurement,transfer
		Type string `json:"type"`
		// Amount
		// required: true
		// example: 5000.00
		Amount float64 `json:"amount"`
		// Currency code
		// required: true
		// example: LYD
		// enum: LYD,USD,EUR,GBP
		Currency string `json:"currency"`
		// Exchange rate (1.0 for local currency)
		// required: true
		// example: 1.0
		ExchangeRate float64 `json:"exchange_rate"`
		// Container ID (cash/bank account)
		// required: true
		// example: 550e8400-e29b-41d4-a716-446655440000
		ContainerID string `json:"container_id"`
		// Project ID
		// required: true
		// example: 550e8400-e29b-41d4-a716-446655440001
		ProjectID string `json:"project_id"`
		// Employee ID (for salary/custody)
		// example: 550e8400-e29b-41d4-a716-446655440002
		EmployeeID string `json:"employee_id,omitempty"`
		// Transaction description
		// required: true
		// example: Monthly grant payment
		Description string `json:"description"`
		// Reference number
		// example: REF-2024-001
		ReferenceNumber string `json:"reference_number,omitempty"`
	}
}

// Transaction response
// swagger:response transactionResponse
type TransactionResponseDoc struct {
	// in: body
	Body struct {
		// Transaction details
		Transaction TransactionInfo `json:"transaction"`
		// Deficit warning flag
		// example: false
		DeficitWarning bool `json:"deficit_warning"`
		// Message
		// example: Transaction completed successfully
		Message string `json:"message,omitempty"`
	}
}

// Transaction information
type TransactionInfo struct {
	// Transaction ID
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// Transaction number
	// example: TXN-2024-000001
	TransactionNumber string `json:"transaction_number"`
	// Transaction type
	// example: income_grant
	Type string `json:"type"`
	// Status
	// example: completed
	// enum: pending,completed,reversed
	Status string `json:"status"`
	// Amount
	// example: 5000.00
	Amount float64 `json:"amount"`
	// Currency
	// example: LYD
	Currency string `json:"currency"`
	// Description
	// example: Monthly grant payment
	Description string `json:"description"`
	// Transaction date
	// example: 2024-01-15T10:30:00Z
	TransactionDate string `json:"transaction_date"`
}

// ============================================================================
// ORGANIZATION/SSA ENDPOINTS
// ============================================================================

// Create organization request
// swagger:parameters createOrganizationRequest
type CreateOrganizationRequestParams struct {
	// Organization data
	// in: body
	// required: true
	Body struct {
		// Organization name
		// required: true
		// example: Acme Corporation
		Name string `json:"name"`
		// Organization name in Arabic
		// example: شركة أكمي
		NameAr string `json:"name_ar,omitempty"`
		// Primary domain
		// required: true
		// example: acme.com
		Domain string `json:"domain"`
		// Contact email
		// required: true
		// example: admin@acme.com
		ContactEmail string `json:"contact_email"`
		// Contact phone
		// example: +218912345678
		ContactPhone string `json:"contact_phone,omitempty"`
	}
}

// Organization response
// swagger:response organizationResponse
type OrganizationResponseDoc struct {
	// in: body
	Body OrganizationInfo
}

// Organization information
type OrganizationInfo struct {
	// Organization ID
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// Organization name
	// example: Acme Corporation
	Name string `json:"name"`
	// Organization name in Arabic
	// example: شركة أكمي
	NameAr string `json:"name_ar,omitempty"`
	// URL slug
	// example: acme-corporation
	Slug string `json:"slug"`
	// Primary domain
	// example: acme.com
	PrimaryDomain string `json:"primary_domain"`
	// Database name
	// example: acme_erp
	DatabaseName string `json:"database_name"`
	// Status
	// example: active
	// enum: pending_verification,active,suspended,deleted
	Status string `json:"status"`
	// Subscription status
	// example: trial
	// enum: trial,active,expired,cancelled
	SubscriptionStatus string `json:"subscription_status"`
}

// Organization stats response
// swagger:response orgStatsResponse
type OrgStatsResponse struct {
	// in: body
	Body struct {
		// Total organizations
		// example: 25
		TotalOrganizations int `json:"total_organizations"`
		// Active organizations
		// example: 20
		ActiveOrganizations int `json:"active_organizations"`
		// Pending verification
		// example: 3
		PendingVerification int `json:"pending_verification"`
		// Suspended organizations
		// example: 2
		SuspendedOrganizations int `json:"suspended_organizations"`
	}
}

// ============================================================================
// REPORT ENDPOINTS
// ============================================================================

// Dashboard summary response
// swagger:response dashboardResponse
type DashboardResponse struct {
	// in: body
	Body struct {
		// Financial summary
		Summary FinancialSummary `json:"summary"`
		// Bank accounts
		BankAccounts []AccountSummary `json:"bank_accounts"`
		// Cash accounts
		CashAccounts []AccountSummary `json:"cash_accounts"`
		// Generated timestamp
		// example: 2024-01-15T10:30:00Z
		GeneratedAt string `json:"generated_at"`
	}
}

// Financial summary
type FinancialSummary struct {
	// Total balance by currency
	TotalBalance []CurrencyAmount `json:"total_balance"`
	// Total income by currency
	TotalIncome []CurrencyAmount `json:"total_income"`
	// Total expenses by currency
	TotalExpenses []CurrencyAmount `json:"total_expenses"`
}

// Currency amount
type CurrencyAmount struct {
	// Currency code
	// example: LYD
	Currency string `json:"currency"`
	// Amount
	// example: 150000.00
	Amount float64 `json:"amount"`
}

// Account summary
type AccountSummary struct {
	// Container ID
	// example: 550e8400-e29b-41d4-a716-446655440000
	ID string `json:"id"`
	// Account name
	// example: Main Bank Account
	Name string `json:"name"`
	// Account type
	// example: bank
	// enum: cash,bank
	Type string `json:"type"`
	// Currency
	// example: LYD
	Currency string `json:"currency"`
	// Current balance
	// example: 75000.00
	Balance float64 `json:"balance"`
}
