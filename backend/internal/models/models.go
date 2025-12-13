package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin       UserRole = "admin"
	RoleFinance     UserRole = "finance"
	RoleAuditor     UserRole = "auditor"
	RoleHR          UserRole = "hr"
	RoleSystemAdmin UserRole = "system_admin" // Super System Admin - hidden role
)

type User struct {
	ID               uuid.UUID  `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	FullName         string     `json:"full_name"`
	FullNameAr       *string    `json:"full_name_ar,omitempty"`
	Role             UserRole   `json:"role"`
	TokenVersion     int        `json:"-"`
	TwoFactorEnabled bool       `json:"two_factor_enabled"`
	IsActive         bool       `json:"is_active"`
	LastLogin        *time.Time `json:"last_login,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	CreatedBy        *uuid.UUID `json:"created_by,omitempty"`
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
	IsLocked    bool         `json:"is_locked"`
	LockedAt    *time.Time   `json:"locked_at,omitempty"`
	LockReason  *string      `json:"lock_reason,omitempty"`
	LockedBy    *uuid.UUID   `json:"locked_by,omitempty"`
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
	EmployeeID            *uuid.UUID        `json:"employee_id,omitempty"`
	MemberID              *uuid.UUID        `json:"member_id,omitempty"`
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
	EmployeeID      *uuid.UUID      `json:"employee_id,omitempty"`
	MemberID        *uuid.UUID      `json:"member_id,omitempty"`
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

type UpdateWidgetConfigRequest struct {
	IsEnabled    *bool    `json:"is_enabled,omitempty"`
	AllowedRoles []string `json:"allowed_roles,omitempty"`
	DisplayOrder *int     `json:"display_order,omitempty"`
}

type WidgetOrderUpdate struct {
	ID           string `json:"id"`
	DisplayOrder int    `json:"display_order"`
}

type DailyPosition struct {
	CashContainers []FinancialContainer `json:"cash_containers"`
	BankContainers []FinancialContainer `json:"bank_containers"`
	GeneratedAt    time.Time            `json:"generated_at"`
}

type CurrencyAmount struct {
	Currency CurrencyCode `json:"currency"`
	Amount   float64      `json:"amount"`
}

type DeficitDetail struct {
	ContainerID   uuid.UUID     `json:"container_id"`
	ContainerName string        `json:"container_name"`
	ContainerType ContainerType `json:"container_type"`
	Currency      CurrencyCode  `json:"currency"`
	Amount        float64       `json:"amount"`
}

type FinancialSummary struct {
	TotalBalance   []CurrencyAmount `json:"total_balance"`
	TotalIncome    []CurrencyAmount `json:"total_income"`
	TotalExpenses  []CurrencyAmount `json:"total_expenses"`
	TotalCustody   []CurrencyAmount `json:"total_custody"`
	TotalDeficit   []CurrencyAmount `json:"total_deficit"`
	DeficitDetails []DeficitDetail  `json:"deficit_details"`
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
	ContainerCurrency CurrencyCode    `json:"container_currency"`
	TransactionDate   time.Time       `json:"transaction_date"`
	CreatedAt         time.Time       `json:"created_at"`
}

type Employee struct {
	ID             uuid.UUID    `json:"id"`
	EmployeeNumber string       `json:"employee_number"`
	FullName       string       `json:"full_name"`
	FullNameAr     *string      `json:"full_name_ar,omitempty"`
	Position       *string      `json:"position,omitempty"`
	PositionAr     *string      `json:"position_ar,omitempty"`
	Department     *string      `json:"department,omitempty"`
	DepartmentAr   *string      `json:"department_ar,omitempty"`
	Phone          *string      `json:"phone,omitempty"`
	Email          *string      `json:"email,omitempty"`
	Salary         *float64     `json:"salary,omitempty"`
	BaseSalary     float64      `json:"base_salary"`
	MonthlyBudget  float64      `json:"monthly_budget"`
	SalaryCurrency CurrencyCode `json:"salary_currency"`
	JoinDate       *time.Time   `json:"join_date,omitempty"`
	IsActive       bool         `json:"is_active"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	CreatedBy      *uuid.UUID   `json:"created_by,omitempty"`
}

type Member struct {
	ID                   uuid.UUID    `json:"id"`
	MembershipNumber     string       `json:"membership_number"`
	FullName             string       `json:"full_name"`
	FullNameAr           *string      `json:"full_name_ar,omitempty"`
	MemberType           string       `json:"member_type"`
	Phone                *string      `json:"phone,omitempty"`
	Email                *string      `json:"email,omitempty"`
	Address              *string      `json:"address,omitempty"`
	AddressAr            *string      `json:"address_ar,omitempty"`
	JoinDate             *time.Time   `json:"join_date,omitempty"`
	Status               string       `json:"status"`
	ValidUntil           *time.Time   `json:"valid_until,omitempty"`
	LastRenewalDate      *time.Time   `json:"last_renewal_date,omitempty"`
	SubscriptionAmount   float64      `json:"subscription_amount"`
	SubscriptionCurrency CurrencyCode `json:"subscription_currency"`
	ComputedStatus       string       `json:"computed_status,omitempty"`
	ExpiringSoon         bool         `json:"expiring_soon"`
	DaysRemaining        int          `json:"days_remaining,omitempty"`
	IsActive             bool         `json:"is_active"`
	CreatedAt            time.Time    `json:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at"`
	CreatedBy            *uuid.UUID   `json:"created_by,omitempty"`
}

type CreateEmployeeRequest struct {
	EmployeeNumber string       `json:"employee_number"`
	FullName       string       `json:"full_name"`
	FullNameAr     *string      `json:"full_name_ar,omitempty"`
	Position       *string      `json:"position,omitempty"`
	PositionAr     *string      `json:"position_ar,omitempty"`
	Department     *string      `json:"department,omitempty"`
	DepartmentAr   *string      `json:"department_ar,omitempty"`
	Phone          *string      `json:"phone,omitempty"`
	Email          *string      `json:"email,omitempty"`
	Salary         *float64     `json:"salary,omitempty"` // Deprecated
	BaseSalary     float64      `json:"base_salary"`
	SalaryCurrency CurrencyCode `json:"salary_currency"`
	JoinDate       *time.Time   `json:"join_date,omitempty"`
}

type UpdateEmployeeRequest struct {
	FullName       *string       `json:"full_name,omitempty"`
	FullNameAr     *string       `json:"full_name_ar,omitempty"`
	Position       *string       `json:"position,omitempty"`
	PositionAr     *string       `json:"position_ar,omitempty"`
	Department     *string       `json:"department,omitempty"`
	DepartmentAr   *string       `json:"department_ar,omitempty"`
	Phone          *string       `json:"phone,omitempty"`
	Email          *string       `json:"email,omitempty"`
	Salary         *float64      `json:"salary,omitempty"` // Deprecated
	BaseSalary     *float64      `json:"base_salary,omitempty"`
	SalaryCurrency *CurrencyCode `json:"salary_currency,omitempty"`
	JoinDate       *time.Time    `json:"join_date,omitempty"`
	IsActive       *bool         `json:"is_active,omitempty"`
}

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
	IsLocked    *bool         `json:"is_locked,omitempty"`
	LockReason  *string       `json:"lock_reason,omitempty"`
}

type CreateTransferRequest struct {
	FromContainerID uuid.UUID    `json:"from_container_id"`
	ToContainerID   uuid.UUID    `json:"to_container_id"`
	Amount          float64      `json:"amount"`
	Currency        CurrencyCode `json:"currency"`
	Description     string       `json:"description"`
	DescriptionAr   *string      `json:"description_ar,omitempty"`
	ReferenceNumber *string      `json:"reference_number,omitempty"`
	TransactionDate *time.Time   `json:"transaction_date,omitempty"`
}

type TransferResponse struct {
	Message           string      `json:"message"`
	DebitTransaction  Transaction `json:"debit_transaction"`
	CreditTransaction Transaction `json:"credit_transaction"`
}

// ============================================================================
// SALARY DEDUCTIONS
// ============================================================================

type SalaryDeduction struct {
	ID         uuid.UUID    `json:"id"`
	EmployeeID uuid.UUID    `json:"employee_id"`
	MonthYear  time.Time    `json:"month_year"`
	Amount     float64      `json:"amount"`
	Currency   CurrencyCode `json:"currency"`
	Reason     string       `json:"reason"`
	ReasonAr   *string      `json:"reason_ar,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	CreatedBy  *uuid.UUID   `json:"created_by,omitempty"`
}

type CreateSalaryDeductionRequest struct {
	EmployeeID uuid.UUID    `json:"employee_id"`
	MonthYear  time.Time    `json:"month_year"`
	Amount     float64      `json:"amount"`
	Currency   CurrencyCode `json:"currency"`
	Reason     string       `json:"reason"`
	ReasonAr   *string      `json:"reason_ar,omitempty"`
}

// ============================================================================
// SALARY BONUSES
// ============================================================================

type SalaryBonus struct {
	ID         uuid.UUID    `json:"id"`
	EmployeeID uuid.UUID    `json:"employee_id"`
	MonthYear  time.Time    `json:"month_year"`
	Amount     float64      `json:"amount"`
	Currency   CurrencyCode `json:"currency"`
	Reason     string       `json:"reason"`
	ReasonAr   *string      `json:"reason_ar,omitempty"`
	CreatedAt  time.Time    `json:"created_at"`
	UpdatedAt  time.Time    `json:"updated_at"`
	CreatedBy  *uuid.UUID   `json:"created_by,omitempty"`
}

type CreateSalaryBonusRequest struct {
	EmployeeID uuid.UUID    `json:"employee_id"`
	MonthYear  time.Time    `json:"month_year"`
	Amount     float64      `json:"amount"`
	Currency   CurrencyCode `json:"currency"`
	Reason     string       `json:"reason"`
	ReasonAr   *string      `json:"reason_ar,omitempty"`
}

type SalarySummary struct {
	EmployeeID      uuid.UUID    `json:"employee_id"`
	EmployeeName    string       `json:"employee_name"`
	EmployeeNumber  string       `json:"employee_number"`
	BaseSalary      float64      `json:"base_salary"`
	TotalBonuses    float64      `json:"total_bonuses"`
	TotalDeductions float64      `json:"total_deductions"`
	NetSalary       float64      `json:"net_salary"`
	Currency        CurrencyCode `json:"currency"`
}

// ============================================================================
// MEMBER SUBSCRIPTIONS
// ============================================================================

type MemberSubscription struct {
	ID            uuid.UUID    `json:"id"`
	MemberID      uuid.UUID    `json:"member_id"`
	TransactionID *uuid.UUID   `json:"transaction_id,omitempty"`
	Amount        float64      `json:"amount"`
	Currency      CurrencyCode `json:"currency"`
	PaymentDate   time.Time    `json:"payment_date"`
	ValidFrom     time.Time    `json:"valid_from"`
	ValidUntil    time.Time    `json:"valid_until"`
	Notes         *string      `json:"notes,omitempty"`
	NotesAr       *string      `json:"notes_ar,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	CreatedBy     *uuid.UUID   `json:"created_by,omitempty"`
}

type CreateMemberSubscriptionRequest struct {
	MemberID      uuid.UUID    `json:"member_id"`
	TransactionID *uuid.UUID   `json:"transaction_id,omitempty"`
	Amount        float64      `json:"amount"`
	Currency      CurrencyCode `json:"currency"`
	Notes         *string      `json:"notes,omitempty"`
	NotesAr       *string      `json:"notes_ar,omitempty"`
}

type RenewMembershipRequest struct {
	MemberID      uuid.UUID    `json:"member_id"`
	Amount        float64      `json:"amount"`
	Currency      CurrencyCode `json:"currency"`
	TransactionID *uuid.UUID   `json:"transaction_id,omitempty"`
	Notes         *string      `json:"notes,omitempty"`
	NotesAr       *string      `json:"notes_ar,omitempty"`
}

// ============================================================================
// EMPLOYEE STATEMENT
// ============================================================================

type EmployeeStatement struct {
	Employee     Employee                    `json:"employee"`
	Transactions []EmployeeTransactionDetail `json:"transactions"`
	Deductions   []SalaryDeduction           `json:"deductions"`
	Summary      EmployeeStatementSummary    `json:"summary"`
	GeneratedAt  time.Time                   `json:"generated_at"`
}

type EmployeeTransactionDetail struct {
	ID                uuid.UUID       `json:"id"`
	TransactionNumber string          `json:"transaction_number"`
	Type              TransactionType `json:"type"`
	Amount            float64         `json:"amount"`
	Currency          CurrencyCode    `json:"currency"`
	Description       string          `json:"description"`
	DescriptionAr     *string         `json:"description_ar,omitempty"`
	ProjectCode       string          `json:"project_code"`
	ProjectName       string          `json:"project_name"`
	ContainerName     string          `json:"container_name"`
	ContainerType     ContainerType   `json:"container_type"`
	TransactionDate   time.Time       `json:"transaction_date"`
	CreatedAt         time.Time       `json:"created_at"`
}

type EmployeeStatementSummary struct {
	TotalSalary      float64 `json:"total_salary"`
	TotalCustody     float64 `json:"total_custody"`
	TotalBonus       float64 `json:"total_bonus"`
	TotalDeductions  float64 `json:"total_deductions"`
	NetPayments      float64 `json:"net_payments"`
	TransactionCount int     `json:"transaction_count"`
}

// ============================================================================
// MEMBER STATEMENT
// ============================================================================

type MemberStatement struct {
	Member        Member                    `json:"member"`
	Subscriptions []MemberSubscription      `json:"subscriptions"`
	Transactions  []MemberTransactionDetail `json:"transactions"`
	Summary       MemberStatementSummary    `json:"summary"`
	GeneratedAt   time.Time                 `json:"generated_at"`
}

type MemberTransactionDetail struct {
	ID                uuid.UUID       `json:"id"`
	TransactionNumber string          `json:"transaction_number"`
	Type              TransactionType `json:"type"`
	Amount            float64         `json:"amount"`
	Currency          CurrencyCode    `json:"currency"`
	Description       string          `json:"description"`
	DescriptionAr     *string         `json:"description_ar,omitempty"`
	ProjectCode       string          `json:"project_code"`
	ProjectName       string          `json:"project_name"`
	ContainerName     string          `json:"container_name"`
	TransactionDate   time.Time       `json:"transaction_date"`
	CreatedAt         time.Time       `json:"created_at"`
}

type MemberStatementSummary struct {
	TotalDeposits     float64 `json:"total_deposits"`
	SubscriptionsPaid int     `json:"subscriptions_paid"`
	MembershipYears   int     `json:"membership_years"`
	TransactionCount  int     `json:"transaction_count"`
}

// ============================================================================
// EXPIRING MEMBERSHIPS (for dashboard widget)
// ============================================================================

type ExpiringMember struct {
	ID               uuid.UUID `json:"id"`
	MembershipNumber string    `json:"membership_number"`
	FullName         string    `json:"full_name"`
	FullNameAr       *string   `json:"full_name_ar,omitempty"`
	MemberType       string    `json:"member_type"`
	Email            *string   `json:"email,omitempty"`
	Phone            *string   `json:"phone,omitempty"`
	ValidUntil       time.Time `json:"valid_until"`
	DaysRemaining    int       `json:"days_remaining"`
}

// ============================================================================
// MULTI-TENANT MODELS (Super System Admin - SSA)
// ============================================================================

// OrganizationStatus represents the status of an organization
type OrganizationStatus string

const (
	OrgStatusPendingVerification OrganizationStatus = "pending_verification"
	OrgStatusActive              OrganizationStatus = "active"
	OrgStatusSuspended           OrganizationStatus = "suspended"
	OrgStatusDeleted             OrganizationStatus = "deleted"
)

// Organization represents a tenant in the multi-tenant system
type Organization struct {
	ID                  uuid.UUID          `json:"id"`
	Name                string             `json:"name"`
	NameAr              *string            `json:"name_ar,omitempty"`
	Slug                string             `json:"slug"`
	PrimaryDomain       string             `json:"primary_domain"`
	DatabaseName        string             `json:"database_name"`
	ContactEmail        string             `json:"contact_email"`
	ContactPhone        *string            `json:"contact_phone,omitempty"`
	Status              OrganizationStatus `json:"status"`
	SubscriptionPlanID  *uuid.UUID         `json:"subscription_plan_id,omitempty"`
	SubscriptionStatus  string             `json:"subscription_status"`
	SubscriptionExpires *time.Time         `json:"subscription_expires_at,omitempty"`
	LogoURL             *string            `json:"logo_url,omitempty"`
	Settings            map[string]any     `json:"settings,omitempty"`
	CreatedAt           time.Time          `json:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at"`
	VerifiedAt          *time.Time         `json:"verified_at,omitempty"`
	SuspendedAt         *time.Time         `json:"suspended_at,omitempty"`
	CreatedBy           uuid.UUID          `json:"created_by"`
}

// AllowedDomain represents a domain allowed for an organization
type AllowedDomain struct {
	ID                 uuid.UUID  `json:"id"`
	OrganizationID     uuid.UUID  `json:"organization_id"`
	Domain             string     `json:"domain"`
	IsPrimary          bool       `json:"is_primary"`
	IsActive           bool       `json:"is_active"`
	Verified           bool       `json:"verified"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty"`
	VerificationMethod *string    `json:"verification_method,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

// DomainVerification represents a pending domain verification
type DomainVerification struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Email          string     `json:"email"`
	HashCode       string     `json:"-"` // Never expose
	ExpiresAt      time.Time  `json:"expires_at"`
	Verified       bool       `json:"verified"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	Attempts       int        `json:"attempts"`
	MaxAttempts    int        `json:"max_attempts"`
	RequestIP      *string    `json:"request_ip,omitempty"`
	VerifyIP       *string    `json:"verify_ip,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// SuperSystemAdmin represents a super system admin user
type SuperSystemAdmin struct {
	ID                  uuid.UUID  `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	FullName            string     `json:"full_name"`
	TwoFactorEnabled    bool       `json:"two_factor_enabled"`
	TwoFactorSecret     *string    `json:"-"`
	TokenVersion        int        `json:"-"`
	LastLogin           *time.Time `json:"last_login,omitempty"`
	LastLoginIP         *string    `json:"last_login_ip,omitempty"`
	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	IsActive            bool       `json:"is_active"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// SubscriptionPlan represents a subscription plan for organizations
type SubscriptionPlan struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	NameAr        *string   `json:"name_ar,omitempty"`
	Code          string    `json:"code"`
	PriceMonthly  float64   `json:"price_monthly"`
	PriceYearly   float64   `json:"price_yearly"`
	Currency      string    `json:"currency"`
	MaxUsers      int       `json:"max_users"`
	MaxStorageGB  int       `json:"max_storage_gb"`
	Features      []string  `json:"features"`
	IsActive      bool      `json:"is_active"`
	IsDefault     bool      `json:"is_default"`
	SortOrder     int       `json:"sort_order"`
	Description   *string   `json:"description,omitempty"`
	DescriptionAr *string   `json:"description_ar,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// SuperAdminAudit represents an audit log entry for SSA actions
type SuperAdminAudit struct {
	ID           uuid.UUID      `json:"id"`
	AdminID      uuid.UUID      `json:"admin_id"`
	AdminEmail   string         `json:"admin_email"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   *uuid.UUID     `json:"resource_id,omitempty"`
	OldValues    map[string]any `json:"old_values,omitempty"`
	NewValues    map[string]any `json:"new_values,omitempty"`
	IPAddress    string         `json:"ip_address"`
	UserAgent    *string        `json:"user_agent,omitempty"`
	RequestID    *string        `json:"request_id,omitempty"`
	PreviousHash *string        `json:"previous_hash,omitempty"`
	CurrentHash  *string        `json:"current_hash,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
}

// ============================================================================
// MULTI-TENANT REQUEST/RESPONSE STRUCTS
// ============================================================================

// CreateOrganizationRequest for creating a new organization
type CreateOrganizationRequest struct {
	Name         string  `json:"name" validate:"required,min=2,max=255"`
	NameAr       *string `json:"name_ar,omitempty"`
	Domain       string  `json:"domain" validate:"required,fqdn"`
	ContactEmail string  `json:"contact_email" validate:"required,email"`
	ContactPhone *string `json:"contact_phone,omitempty"`
}

// VerifyDomainRequest for verifying domain ownership
type VerifyDomainRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	Code           string    `json:"code" validate:"required,len=6"`
}

// CreateTenantAdminRequest for creating first admin of organization
type CreateTenantAdminRequest struct {
	Username   string  `json:"username" validate:"required,min=3,max=50"`
	Email      string  `json:"email" validate:"required,email"`
	FullName   string  `json:"full_name" validate:"required,min=2,max=255"`
	FullNameAr *string `json:"full_name_ar,omitempty"`
}

// CreateTenantAdminResponse returned after creating tenant admin
type CreateTenantAdminResponse struct {
	Username          string `json:"username"`
	Email             string `json:"email"`
	TemporaryPassword string `json:"temporary_password"`
	OrganizationName  string `json:"organization_name"`
	LoginURL          string `json:"login_url"`
}

// ResendVerificationRequest for resending verification email
type ResendVerificationRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	Email          string    `json:"email" validate:"required,email"`
}

// SuspendOrganizationRequest for suspending an organization
type SuspendOrganizationRequest struct {
	Reason   string  `json:"reason" validate:"required"`
	ReasonAr *string `json:"reason_ar,omitempty"`
}

// OrganizationListItem for listing organizations
type OrganizationListItem struct {
	ID            uuid.UUID          `json:"id"`
	Name          string             `json:"name"`
	NameAr        *string            `json:"name_ar,omitempty"`
	PrimaryDomain string             `json:"primary_domain"`
	Status        OrganizationStatus `json:"status"`
	UserCount     int                `json:"user_count"`
	LastActivity  *time.Time         `json:"last_activity,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
}

// OrganizationStats for SSA dashboard
type OrganizationStats struct {
	TotalOrganizations     int     `json:"total_organizations"`
	ActiveOrganizations    int     `json:"active_organizations"`
	PendingVerification    int     `json:"pending_verification"`
	SuspendedOrganizations int     `json:"suspended_organizations"`
	TotalUsers             int     `json:"total_users"`
	TotalStorageUsedGB     float64 `json:"total_storage_used_gb"`
}

// SSALoginRequest for super system admin login
type SSALoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// SSALoginResponse for super system admin login response
type SSALoginResponse struct {
	Admin            SuperSystemAdmin `json:"admin"`
	Token            string           `json:"token"`
	Requires2FA      bool             `json:"requires_2fa"`
	PendingSessionID *string          `json:"pending_session_id,omitempty"`
}

// SSAVerify2FARequest for 2FA verification
type SSAVerify2FARequest struct {
	SessionID string `json:"session_id" validate:"required"`
	Code      string `json:"code" validate:"required,len=6"`
}

// ============================================================================
// ORGANIZATION-SPECIFIC MODELS (Tenant Database)
// ============================================================================

// EmailTemplate for customizable email templates per organization
type EmailTemplate struct {
	ID          uuid.UUID  `json:"id"`
	TemplateKey string     `json:"template_key"`
	Name        string     `json:"name"`
	NameAr      *string    `json:"name_ar,omitempty"`
	Subject     string     `json:"subject"`
	SubjectAr   *string    `json:"subject_ar,omitempty"`
	BodyHTML    string     `json:"body_html"`
	BodyText    *string    `json:"body_text,omitempty"`
	Description *string    `json:"description,omitempty"`
	Variables   []string   `json:"variables"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
}

// SMTPSettings for organization-specific SMTP configuration
type SMTPSettings struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	SMTPHost          string     `json:"smtp_host"`
	SMTPPort          int        `json:"smtp_port"`
	Username          string     `json:"username,omitempty"`
	PasswordEncrypted string     `json:"-"`
	FromAddress       string     `json:"from_address"`
	FromName          *string    `json:"from_name,omitempty"`
	UseTLS            bool       `json:"use_tls"`
	IsActive          bool       `json:"is_active"`
	IsDefault         bool       `json:"is_default"`
	TestEmailSentAt   *time.Time `json:"test_email_sent_at,omitempty"`
	LastError         *string    `json:"last_error,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	CreatedBy         *uuid.UUID `json:"created_by,omitempty"`
}

// OrganizationSettings key-value store for tenant-specific settings
type OrganizationSettings struct {
	ID           uuid.UUID `json:"id"`
	SettingKey   string    `json:"setting_key"`
	SettingValue *string   `json:"setting_value,omitempty"`
	SettingType  string    `json:"setting_type"`
	Description  *string   `json:"description,omitempty"`
	IsEncrypted  bool      `json:"is_encrypted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ============================================================================
// EMAIL TEMPLATE REQUEST/RESPONSE STRUCTS
// ============================================================================

// UpdateEmailTemplateRequest for updating email template
type UpdateEmailTemplateRequest struct {
	Name        *string  `json:"name,omitempty"`
	NameAr      *string  `json:"name_ar,omitempty"`
	Subject     *string  `json:"subject,omitempty"`
	SubjectAr   *string  `json:"subject_ar,omitempty"`
	BodyHTML    *string  `json:"body_html,omitempty"`
	BodyText    *string  `json:"body_text,omitempty"`
	Description *string  `json:"description,omitempty"`
	Variables   []string `json:"variables,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
}

// CreateSMTPSettingsRequest for creating SMTP settings
type CreateSMTPSettingsRequest struct {
	Name        string  `json:"name" validate:"required"`
	SMTPHost    string  `json:"smtp_host" validate:"required"`
	SMTPPort    int     `json:"smtp_port" validate:"required,min=1,max=65535"`
	Username    string  `json:"username"`
	Password    string  `json:"password"`
	FromAddress string  `json:"from_address" validate:"required,email"`
	FromName    *string `json:"from_name,omitempty"`
	UseTLS      bool    `json:"use_tls"`
	IsDefault   bool    `json:"is_default"`
}

// UpdateSMTPSettingsRequest for updating SMTP settings
type UpdateSMTPSettingsRequest struct {
	Name        *string `json:"name,omitempty"`
	SMTPHost    *string `json:"smtp_host,omitempty"`
	SMTPPort    *int    `json:"smtp_port,omitempty"`
	Username    *string `json:"username,omitempty"`
	Password    *string `json:"password,omitempty"`
	FromAddress *string `json:"from_address,omitempty"`
	FromName    *string `json:"from_name,omitempty"`
	UseTLS      *bool   `json:"use_tls,omitempty"`
	IsDefault   *bool   `json:"is_default,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// TestSMTPRequest for testing SMTP connection
type TestSMTPRequest struct {
	ToEmail string `json:"to_email" validate:"required,email"`
}

// TemplatePreviewRequest for previewing email template with variables
type TemplatePreviewRequest struct {
	TemplateKey string            `json:"template_key" validate:"required"`
	Variables   map[string]string `json:"variables"`
}

// TemplatePreviewResponse for template preview response
type TemplatePreviewResponse struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
