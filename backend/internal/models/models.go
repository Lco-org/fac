package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleFinance UserRole = "finance"
	RoleAuditor UserRole = "auditor"
)

type User struct {
	ID           uuid.UUID  `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"`
	FullName     string     `json:"full_name"`
	FullNameAr   *string    `json:"full_name_ar,omitempty"`
	Role         UserRole   `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLogin    *time.Time `json:"last_login,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty"`
}

type ContainerType string

const (
	ContainerCash ContainerType = "cash"
	ContainerBank ContainerType = "bank"
)

type CurrencyCode string

const (
	CurrencyLYD CurrencyCode = "LYD"
	CurrencyUSD CurrencyCode = "USD"
	CurrencyEUR CurrencyCode = "EUR"
	CurrencyGBP CurrencyCode = "GBP"
)

type FinancialContainer struct {
	ID            uuid.UUID     `json:"id"`
	Name          string        `json:"name"`
	NameAr        *string       `json:"name_ar,omitempty"`
	Type          ContainerType `json:"type"`
	Currency      CurrencyCode  `json:"currency"`
	Balance       float64       `json:"balance"`
	BankName      *string       `json:"bank_name,omitempty"`
	BankNameAr    *string       `json:"bank_name_ar,omitempty"`
	AccountNumber *string       `json:"account_number,omitempty"`
	IBAN          *string       `json:"iban,omitempty"`
	SwiftCode     *string       `json:"swift_code,omitempty"`
	Location      *string       `json:"location,omitempty"`
	LocationAr    *string       `json:"location_ar,omitempty"`
	IsActive      bool          `json:"is_active"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	CreatedBy     *uuid.UUID    `json:"created_by,omitempty"`
}

type Project struct {
	ID          uuid.UUID    `json:"id"`
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	NameAr      *string      `json:"name_ar,omitempty"`
	Description *string      `json:"description,omitempty"`
	DonorName   *string      `json:"donor_name,omitempty"`
	DonorNameAr *string      `json:"donor_name_ar,omitempty"`
	StartDate   *time.Time   `json:"start_date,omitempty"`
	EndDate     *time.Time   `json:"end_date,omitempty"`
	Budget      *float64     `json:"budget,omitempty"`
	Currency    CurrencyCode `json:"currency"`
	IsActive    bool         `json:"is_active"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CreatedBy   *uuid.UUID   `json:"created_by,omitempty"`
}

type TransactionType string

const (
	TransactionIncomeGrant             TransactionType = "income_grant"
	TransactionIncomeDonation          TransactionType = "income_donation"
	TransactionDisbursementCustody     TransactionType = "disbursement_custody"
	TransactionDisbursementSalary      TransactionType = "disbursement_salary"
	TransactionDisbursementOperational TransactionType = "disbursement_operational"
	TransactionDisbursementDonation    TransactionType = "disbursement_donation"
	TransactionDisbursementBonus       TransactionType = "disbursement_bonus"
	TransactionProcurement             TransactionType = "procurement"
	TransactionTransfer                TransactionType = "transfer"
)

type TransactionStatus string

const (
	TransactionPending   TransactionStatus = "pending"
	TransactionCompleted TransactionStatus = "completed"
	TransactionReversed  TransactionStatus = "reversed"
)

type Transaction struct {
	ID                    uuid.UUID         `json:"id"`
	TransactionNumber     string            `json:"transaction_number"`
	Type                  TransactionType   `json:"type"`
	Status                TransactionStatus `json:"status"`
	Amount                float64           `json:"amount"`
	Currency              CurrencyCode      `json:"currency"`
	ExchangeRate          float64           `json:"exchange_rate"`
	AmountLocal           float64           `json:"amount_local"`
	ContainerID           uuid.UUID         `json:"container_id"`
	ProjectID             uuid.UUID         `json:"project_id"`
	Description           string            `json:"description"`
	DescriptionAr         *string           `json:"description_ar,omitempty"`
	PayeePayer            *string           `json:"payee_payer,omitempty"`
	PayeePayerAr          *string           `json:"payee_payer_ar,omitempty"`
	ReferenceNumber       *string           `json:"reference_number,omitempty"`
	IsDeficit             bool              `json:"is_deficit"`
	DeficitApprovedBy     *uuid.UUID        `json:"deficit_approved_by,omitempty"`
	DeficitWarningShown   bool              `json:"deficit_warning_shown"`
	BalanceBefore         *float64          `json:"balance_before,omitempty"`
	BalanceAfter          *float64          `json:"balance_after,omitempty"`
	AttachmentURL         *string           `json:"attachment_url,omitempty"`
	TransactionDate       time.Time         `json:"transaction_date"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
	CreatedBy             uuid.UUID         `json:"created_by"`
	ReversedAt            *time.Time        `json:"reversed_at,omitempty"`
	ReversedBy            *uuid.UUID        `json:"reversed_by,omitempty"`
	ReversalReason        *string           `json:"reversal_reason,omitempty"`
	OriginalTransactionID *uuid.UUID        `json:"original_transaction_id,omitempty"`
}

// Request/Response DTOs

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type CreateUserRequest struct {
	Username   string   `json:"username"`
	Email      string   `json:"email"`
	Password   string   `json:"password"`
	FullName   string   `json:"full_name"`
	FullNameAr *string  `json:"full_name_ar,omitempty"`
	Role       UserRole `json:"role"`
}

type UpdateUserRequest struct {
	Email      *string `json:"email,omitempty"`
	FullName   *string `json:"full_name,omitempty"`
	FullNameAr *string `json:"full_name_ar,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
}

type CreateTransactionRequest struct {
	Type            TransactionType `json:"type"`
	Amount          float64         `json:"amount"`
	Currency        CurrencyCode    `json:"currency"`
	ExchangeRate    float64         `json:"exchange_rate"`
	ContainerID     uuid.UUID       `json:"container_id"`
	ProjectID       uuid.UUID       `json:"project_id"`
	Description     string          `json:"description"`
	DescriptionAr   *string         `json:"description_ar,omitempty"`
	PayeePayer      *string         `json:"payee_payer,omitempty"`
	PayeePayerAr    *string         `json:"payee_payer_ar,omitempty"`
	ReferenceNumber *string         `json:"reference_number,omitempty"`
	TransactionDate *time.Time      `json:"transaction_date,omitempty"`
	ConfirmDeficit  bool            `json:"confirm_deficit"`
}

type TransactionResponse struct {
	Transaction      Transaction `json:"transaction"`
	DeficitWarning   bool        `json:"deficit_warning"`
	RequiresApproval bool        `json:"requires_approval"`
	Message          string      `json:"message,omitempty"`
}

// ContainerStatement represents detailed financial statement for a container
type ContainerStatement struct {
	Container    FinancialContainer        `json:"container"`
	Transactions []TransactionDetail       `json:"transactions"`
	Summary      ContainerStatementSummary `json:"summary"`
	GeneratedAt  time.Time                 `json:"generated_at"`
}

type ContainerStatementSummary struct {
	OpeningBalance   float64 `json:"opening_balance"`
	TotalIncome      float64 `json:"total_income"`
	TotalExpenses    float64 `json:"total_expenses"`
	ClosingBalance   float64 `json:"closing_balance"`
	TransactionCount int     `json:"transaction_count"`
}

type TransactionDetail struct {
	ID                uuid.UUID       `json:"id"`
	TransactionNumber string          `json:"transaction_number"`
	Type              TransactionType `json:"type"`
	Amount            float64         `json:"amount"`
	Currency          CurrencyCode    `json:"currency"`
	Description       string          `json:"description"`
	DescriptionAr     *string         `json:"description_ar,omitempty"`
	PayeePayer        *string         `json:"payee_payer,omitempty"`
	ProjectCode       string          `json:"project_code"`
	ProjectName       string          `json:"project_name"`
	BalanceBefore     float64         `json:"balance_before"`
	BalanceAfter      float64         `json:"balance_after"`
	TransactionDate   time.Time       `json:"transaction_date"`
	CreatedAt         time.Time       `json:"created_at"`
}

// ProjectStatement represents detailed financial statement for a project
type ProjectStatement struct {
	Project      Project                    `json:"project"`
	Transactions []ProjectTransactionDetail `json:"transactions"`
	Summary      ProjectStatementSummary    `json:"summary"`
	GeneratedAt  time.Time                  `json:"generated_at"`
}

type ProjectStatementSummary struct {
	TotalIncome       float64 `json:"total_income"`
	TotalExpenses     float64 `json:"total_expenses"`
	NetPosition       float64 `json:"net_position"`
	TransactionCount  int     `json:"transaction_count"`
	Budget            float64 `json:"budget"`
	BudgetUtilization float64 `json:"budget_utilization"`
}

type DashboardSummary struct {
	Summary        FinancialSummary           `json:"summary"`
	BankAccounts   []BankDashboardSummary     `json:"bank_accounts"`
	CashAccounts   []BankDashboardSummary     `json:"cash_accounts"`
	Projects       []ProjectDashboardSummary  `json:"projects"`
	RecentIncome   []ProjectTransactionDetail `json:"recent_income"`
	RecentExpenses []ProjectTransactionDetail `json:"recent_expenses"`
	GeneratedAt    time.Time                  `json:"generated_at"`
}

type BankDashboardSummary struct {
	Container        FinancialContainer  `json:"container"`
	LastTransactions []TransactionDetail `json:"last_transactions"`
}

type ProjectDashboardSummary struct {
	Project          Project                    `json:"project"`
	LastTransactions []ProjectTransactionDetail `json:"last_transactions"`
}

// DashboardWidget represents a configurable dashboard widget
type DashboardWidget struct {
	ID           uuid.UUID `json:"id"`
	WidgetKey    string    `json:"widget_key"`
	WidgetName   string    `json:"widget_name"`
	WidgetNameEn string    `json:"widget_name_en"`
	IsEnabled    bool      `json:"is_enabled"`
	AllowedRoles []string  `json:"allowed_roles"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpdateWidgetConfigRequest for updating widget configuration
type UpdateWidgetConfigRequest struct {
	IsEnabled    *bool    `json:"is_enabled,omitempty"`
	AllowedRoles []string `json:"allowed_roles,omitempty"`
	DisplayOrder *int     `json:"display_order,omitempty"`
}

type DailyPosition struct {
	CashContainers []FinancialContainer `json:"cash_containers"`
	BankContainers []FinancialContainer `json:"bank_containers"`
	GeneratedAt    time.Time            `json:"generated_at"`
}

// CurrencyAmount represents an amount in a specific currency
type CurrencyAmount struct {
	Currency CurrencyCode `json:"currency"`
	Amount   float64      `json:"amount"`
}

// FinancialSummary represents organization-wide financial totals
type FinancialSummary struct {
	TotalBalance  []CurrencyAmount `json:"total_balance"`
	TotalIncome   []CurrencyAmount `json:"total_income"`
	TotalExpenses []CurrencyAmount `json:"total_expenses"`
	TotalCustody  []CurrencyAmount `json:"total_custody"`
}

type ProjectTransactionDetail struct {
	ID                uuid.UUID       `json:"id"`
	TransactionNumber string          `json:"transaction_number"`
	Type              TransactionType `json:"type"`
	Amount            float64         `json:"amount"`
	Currency          CurrencyCode    `json:"currency"`
	Description       string          `json:"description"`
	DescriptionAr     *string         `json:"description_ar,omitempty"`
	PayeePayer        *string         `json:"payee_payer,omitempty"`
	ProjectCode       string          `json:"project_code,omitempty"`
	ProjectName       string          `json:"project_name,omitempty"`
	ContainerName     string          `json:"container_name"`
	ContainerType     ContainerType   `json:"container_type"`
	TransactionDate   time.Time       `json:"transaction_date"`
	CreatedAt         time.Time       `json:"created_at"`
}

// Employee represents an organization employee
type Employee struct {
	ID             uuid.UUID  `json:"id"`
	EmployeeNumber string     `json:"employee_number"`
	FullName       string     `json:"full_name"`
	FullNameAr     *string    `json:"full_name_ar,omitempty"`
	Position       *string    `json:"position,omitempty"`
	PositionAr     *string    `json:"position_ar,omitempty"`
	Department     *string    `json:"department,omitempty"`
	DepartmentAr   *string    `json:"department_ar,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	Email          *string    `json:"email,omitempty"`
	Salary         *float64   `json:"salary,omitempty"`
	JoinDate       *time.Time `json:"join_date,omitempty"`
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
}

// Member represents an affiliated member
type Member struct {
	ID               uuid.UUID  `json:"id"`
	MembershipNumber string     `json:"membership_number"`
	FullName         string     `json:"full_name"`
	FullNameAr       *string    `json:"full_name_ar,omitempty"`
	MemberType       string     `json:"member_type"`
	Phone            *string    `json:"phone,omitempty"`
	Email            *string    `json:"email,omitempty"`
	Address          *string    `json:"address,omitempty"`
	AddressAr        *string    `json:"address_ar,omitempty"`
	JoinDate         *time.Time `json:"join_date,omitempty"`
	Status           string     `json:"status"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CreatedBy        *uuid.UUID `json:"created_by,omitempty"`
}

// CreateEmployeeRequest for creating new employees
type CreateEmployeeRequest struct {
	EmployeeNumber string     `json:"employee_number"`
	FullName       string     `json:"full_name"`
	FullNameAr     *string    `json:"full_name_ar,omitempty"`
	Position       *string    `json:"position,omitempty"`
	PositionAr     *string    `json:"position_ar,omitempty"`
	Department     *string    `json:"department,omitempty"`
	DepartmentAr   *string    `json:"department_ar,omitempty"`
	Phone          *string    `json:"phone,omitempty"`
	Email          *string    `json:"email,omitempty"`
	Salary         *float64   `json:"salary,omitempty"`
	JoinDate       *time.Time `json:"join_date,omitempty"`
}

// UpdateEmployeeRequest for updating employees
type UpdateEmployeeRequest struct {
	FullName     *string    `json:"full_name,omitempty"`
	FullNameAr   *string    `json:"full_name_ar,omitempty"`
	Position     *string    `json:"position,omitempty"`
	PositionAr   *string    `json:"position_ar,omitempty"`
	Department   *string    `json:"department,omitempty"`
	DepartmentAr *string    `json:"department_ar,omitempty"`
	Phone        *string    `json:"phone,omitempty"`
	Email        *string    `json:"email,omitempty"`
	Salary       *float64   `json:"salary,omitempty"`
	JoinDate     *time.Time `json:"join_date,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
}

// CreateMemberRequest for creating new members
type CreateMemberRequest struct {
	MembershipNumber string     `json:"membership_number"`
	FullName         string     `json:"full_name"`
	FullNameAr       *string    `json:"full_name_ar,omitempty"`
	MemberType       string     `json:"member_type"`
	Phone            *string    `json:"phone,omitempty"`
	Email            *string    `json:"email,omitempty"`
	Address          *string    `json:"address,omitempty"`
	AddressAr        *string    `json:"address_ar,omitempty"`
	JoinDate         *time.Time `json:"join_date,omitempty"`
}

// UpdateMemberRequest for updating members
type UpdateMemberRequest struct {
	FullName   *string    `json:"full_name,omitempty"`
	FullNameAr *string    `json:"full_name_ar,omitempty"`
	MemberType *string    `json:"member_type,omitempty"`
	Phone      *string    `json:"phone,omitempty"`
	Email      *string    `json:"email,omitempty"`
	Address    *string    `json:"address,omitempty"`
	AddressAr  *string    `json:"address_ar,omitempty"`
	JoinDate   *time.Time `json:"join_date,omitempty"`
	Status     *string    `json:"status,omitempty"`
	IsActive   *bool      `json:"is_active,omitempty"`
}

// CreateProjectRequest for creating new projects
type CreateProjectRequest struct {
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	NameAr      *string      `json:"name_ar,omitempty"`
	Description *string      `json:"description,omitempty"`
	DonorName   *string      `json:"donor_name,omitempty"`
	DonorNameAr *string      `json:"donor_name_ar,omitempty"`
	StartDate   *time.Time   `json:"start_date,omitempty"`
	EndDate     *time.Time   `json:"end_date,omitempty"`
	Budget      *float64     `json:"budget,omitempty"`
	Currency    CurrencyCode `json:"currency"`
}

// UpdateProjectRequest for updating projects
type UpdateProjectRequest struct {
	Name        *string       `json:"name,omitempty"`
	NameAr      *string       `json:"name_ar,omitempty"`
	Description *string       `json:"description,omitempty"`
	DonorName   *string       `json:"donor_name,omitempty"`
	DonorNameAr *string       `json:"donor_name_ar,omitempty"`
	StartDate   *time.Time    `json:"start_date,omitempty"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
	Budget      *float64      `json:"budget,omitempty"`
	Currency    *CurrencyCode `json:"currency,omitempty"`
	IsActive    *bool         `json:"is_active,omitempty"`
}
