package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TransferService struct {
	db *database.DB
}

func NewTransferService(db *database.DB) *TransferService {
	return &TransferService{db: db}
}

// CreateTransfer creates a transfer between two containers
func (s *TransferService) CreateTransfer(ctx context.Context, req *models.CreateTransferRequest, userID uuid.UUID) (*models.TransferResponse, error) {
	// Validate request
	if req.Amount <= 0 {
		return nil, fmt.Errorf("amount must be greater than 0")
	}

	if req.FromContainerID == req.ToContainerID {
		return nil, fmt.Errorf("source and destination containers must be different")
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get source container
	var fromContainer models.FinancialContainer
	err = tx.QueryRowContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, is_active
		FROM financial_containers
		WHERE id = $1
	`, req.FromContainerID).Scan(
		&fromContainer.ID,
		&fromContainer.Name,
		&fromContainer.NameAr,
		&fromContainer.Type,
		&fromContainer.Currency,
		&fromContainer.Balance,
		&fromContainer.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, ErrContainerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch source container: %w", err)
	}

	if !fromContainer.IsActive {
		return nil, fmt.Errorf("source container is not active")
	}

	// Get destination container
	var toContainer models.FinancialContainer
	err = tx.QueryRowContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, is_active
		FROM financial_containers
		WHERE id = $1
	`, req.ToContainerID).Scan(
		&toContainer.ID,
		&toContainer.Name,
		&toContainer.NameAr,
		&toContainer.Type,
		&toContainer.Currency,
		&toContainer.Balance,
		&toContainer.IsActive,
	)
	if err == sql.ErrNoRows {
		return nil, ErrContainerNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch destination container: %w", err)
	}

	if !toContainer.IsActive {
		return nil, fmt.Errorf("destination container is not active")
	}

	// Validate currency match
	if fromContainer.Currency != toContainer.Currency {
		return nil, fmt.Errorf("currency mismatch: source container is %s but destination is %s", fromContainer.Currency, toContainer.Currency)
	}

	// Validate currency in request matches containers
	if req.Currency != fromContainer.Currency {
		return nil, fmt.Errorf("request currency %s does not match container currency %s", req.Currency, fromContainer.Currency)
	}

	// Check sufficient balance
	if fromContainer.Balance < req.Amount {
		return nil, ErrInsufficientBalance
	}

	// Generate transaction number (simplified - consider using a sequence)
	transactionDate := time.Now()
	if req.TransactionDate != nil {
		transactionDate = *req.TransactionDate
	}

	// Create debit transaction (withdrawal from source)
	debitID := uuid.New()
	debitNumber := fmt.Sprintf("TRF-OUT-%s", debitID.String()[:8])

	balanceBeforeDebit := fromContainer.Balance
	balanceAfterDebit := fromContainer.Balance - req.Amount

	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactions (
			id, transaction_number, type, status, amount, currency, exchange_rate, amount_local,
			container_id, project_id, description, description_ar, reference_number,
			is_deficit, deficit_warning_shown, balance_before, balance_after,
			transaction_date, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`,
		debitID,
		debitNumber,
		models.TransactionTransfer,
		models.TransactionCompleted,
		req.Amount,
		req.Currency,
		1.0,
		req.Amount,
		req.FromContainerID,
		nil, // No project for transfers
		req.Description,
		req.DescriptionAr,
		req.ReferenceNumber,
		false,
		false,
		balanceBeforeDebit,
		balanceAfterDebit,
		transactionDate,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create debit transaction: %w", err)
	}

	// Create credit transaction (deposit to destination)
	creditID := uuid.New()
	creditNumber := fmt.Sprintf("TRF-IN-%s", creditID.String()[:8])

	balanceBeforeCredit := toContainer.Balance
	balanceAfterCredit := toContainer.Balance + req.Amount

	_, err = tx.ExecContext(ctx, `
		INSERT INTO transactions (
			id, transaction_number, type, status, amount, currency, exchange_rate, amount_local,
			container_id, project_id, description, description_ar, reference_number,
			is_deficit, deficit_warning_shown, balance_before, balance_after,
			transaction_date, created_by, original_transaction_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)
	`,
		creditID,
		creditNumber,
		models.TransactionTransfer,
		models.TransactionCompleted,
		req.Amount,
		req.Currency,
		1.0,
		req.Amount,
		req.ToContainerID,
		nil, // No project for transfers
		req.Description,
		req.DescriptionAr,
		req.ReferenceNumber,
		false,
		false,
		balanceBeforeCredit,
		balanceAfterCredit,
		transactionDate,
		userID,
		debitID, // Link to debit transaction
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit transaction: %w", err)
	}

	// Update source container balance
	_, err = tx.ExecContext(ctx, `
		UPDATE financial_containers
		SET balance = balance - $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, req.Amount, req.FromContainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to update source container balance: %w", err)
	}

	// Update destination container balance
	_, err = tx.ExecContext(ctx, `
		UPDATE financial_containers
		SET balance = balance + $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, req.Amount, req.ToContainerID)
	if err != nil {
		return nil, fmt.Errorf("failed to update destination container balance: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transfer: %w", err)
	}

	// Fetch the created transactions
	debitTx, err := s.getTransactionByID(ctx, debitID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch debit transaction: %w", err)
	}

	creditTx, err := s.getTransactionByID(ctx, creditID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch credit transaction: %w", err)
	}

	fromName := fromContainer.Name
	if fromContainer.NameAr != nil && *fromContainer.NameAr != "" {
		fromName = *fromContainer.NameAr
	}

	toName := toContainer.Name
	if toContainer.NameAr != nil && *toContainer.NameAr != "" {
		toName = *toContainer.NameAr
	}

	message := fmt.Sprintf("تم التحويل بنجاح من %s إلى %s", fromName, toName)

	return &models.TransferResponse{
		DebitTransaction:  *debitTx,
		CreditTransaction: *creditTx,
		Message:           message,
	}, nil
}

// getTransactionByID retrieves a transaction by ID
func (s *TransferService) getTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := s.db.QueryRowContext(ctx, `
		SELECT 
			id, transaction_number, type, status, amount, currency, exchange_rate, amount_local,
			container_id, project_id, description, description_ar, payee_payer, payee_payer_ar,
			reference_number, is_deficit, deficit_approved_by, deficit_warning_shown,
			balance_before, balance_after, attachment_url, transaction_date,
			created_at, updated_at, created_by, reversed_at, reversed_by, reversal_reason, original_transaction_id
		FROM transactions
		WHERE id = $1
	`, id).Scan(
		&tx.ID,
		&tx.TransactionNumber,
		&tx.Type,
		&tx.Status,
		&tx.Amount,
		&tx.Currency,
		&tx.ExchangeRate,
		&tx.AmountLocal,
		&tx.ContainerID,
		&tx.ProjectID,
		&tx.Description,
		&tx.DescriptionAr,
		&tx.PayeePayer,
		&tx.PayeePayerAr,
		&tx.ReferenceNumber,
		&tx.IsDeficit,
		&tx.DeficitApprovedBy,
		&tx.DeficitWarningShown,
		&tx.BalanceBefore,
		&tx.BalanceAfter,
		&tx.AttachmentURL,
		&tx.TransactionDate,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&tx.CreatedBy,
		&tx.ReversedAt,
		&tx.ReversedBy,
		&tx.ReversalReason,
		&tx.OriginalTransactionID,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("transaction not found")
	}
	if err != nil {
		return nil, err
	}
	return &tx, nil
}
