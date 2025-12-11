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
	ID               uuid.UUID  `json:"id"`
	Username         string     `json:"username"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"-"`
	FullName         string     `json:"full_name"`
	FullNameAr       *string    `json:"full_name_ar,omitempty"`
	Role             UserRole   `json:"role"`
	Permissions      []string   `json:"permissions"`
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
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Password    string   `json:"password"`
	FullName    string   `json:"full_name"`
	FullNameAr  *string  `json:"full_name_ar,omitempty"`
	Role        UserRole `json:"role"`
	Permissions []string `json:"permissions"`
}

type UpdateUserRequest struct {
	Email       *string  `json:"email,omitempty"`
	FullName    *string  `json:"full_name,omitempty"`
	FullNameAr  *string  `json:"full_name_ar,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
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
