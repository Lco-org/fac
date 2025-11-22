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

type DailyPositionReport struct {
	CashContainers []ContainerPosition `json:"cash_containers"`
	BankContainers []ContainerPosition `json:"bank_containers"`
	TotalCash      float64             `json:"total_cash"`
	TotalBank      float64             `json:"total_bank"`
	TotalDeficit   float64             `json:"total_deficit"`
	GeneratedAt    time.Time           `json:"generated_at"`
}

type ContainerPosition struct {
	ID            uuid.UUID     `json:"id"`
	Name          string        `json:"name"`
	NameAr        *string       `json:"name_ar,omitempty"`
	Type          ContainerType `json:"type"`
	Currency      CurrencyCode  `json:"currency"`
	Balance       float64       `json:"balance"`
	BalanceStatus string        `json:"balance_status"`
	TodayIncome   float64       `json:"today_income"`
	TodayExpenses float64       `json:"today_expenses"`
	BankName      *string       `json:"bank_name,omitempty"`
	AccountNumber *string       `json:"account_number,omitempty"`
	Location      *string       `json:"location,omitempty"`
}

type ProjectSummary struct {
	ID               uuid.UUID    `json:"id"`
	Code             string       `json:"code"`
	Name             string       `json:"name"`
	NameAr           *string      `json:"name_ar,omitempty"`
	Currency         CurrencyCode `json:"currency"`
	DonorName        *string      `json:"donor_name,omitempty"`
	TotalIncome      float64      `json:"total_income"`
	TotalExpenses    float64      `json:"total_expenses"`
	NetPosition      float64      `json:"net_position"`
	TransactionCount int          `json:"transaction_count"`
}

type DeficitReport struct {
	Containers      []DeficitContainer `json:"containers"`
	TotalDeficit    float64            `json:"total_deficit"`
	ContainersCount int                `json:"containers_count"`
	GeneratedAt     time.Time          `json:"generated_at"`
}

type DeficitContainer struct {
	ID                       uuid.UUID     `json:"id"`
	Name                     string        `json:"name"`
	NameAr                   *string       `json:"name_ar,omitempty"`
	Type                     ContainerType `json:"type"`
	Currency                 CurrencyCode  `json:"currency"`
	DeficitAmount            float64       `json:"deficit_amount"`
	DeficitTransactionsCount int           `json:"deficit_transactions_count"`
	FirstDeficitDate         *time.Time    `json:"first_deficit_date,omitempty"`
	LastDeficitDate          *time.Time    `json:"last_deficit_date,omitempty"`
}
