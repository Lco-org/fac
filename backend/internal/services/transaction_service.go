package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TransactionService struct {
	db *database.DB
}

func NewTransactionService(db *database.DB) *TransactionService {
	return &TransactionService{db: db}
}

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrContainerNotFound   = errors.New("container not found")
	ErrProjectNotFound     = errors.New("project not found")
	ErrInvalidAmount       = errors.New("invalid amount")
)

// CreateTransaction handles the complete transaction creation with deficit handling
func (s *TransactionService) CreateTransaction(ctx context.Context, req *models.CreateTransactionRequest, userID uuid.UUID) (*models.TransactionResponse, error) {
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}

	var response models.TransactionResponse
	var transaction models.Transaction

	// Execute within database transaction
	err := s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// 1. Get current container balance (with row lock)
		var container models.FinancialContainer
		err := tx.QueryRowContext(ctx, `
			SELECT id, name, type, currency, balance, is_active
			FROM financial_containers
			WHERE id = $1 AND is_active = TRUE
			FOR UPDATE
		`, req.ContainerID).Scan(
			&container.ID,
			&container.Name,
			&container.Type,
			&container.Currency,
			&container.Balance,
			&container.IsActive,
		)
		if err == sql.ErrNoRows {
			return ErrContainerNotFound
		}
		if err != nil {
			return fmt.Errorf("failed to get container: %w", err)
		}

		// 2. Verify project exists
		var projectExists bool
		err = tx.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND is_active = TRUE)
		`, req.ProjectID).Scan(&projectExists)
		if err != nil {
			return fmt.Errorf("failed to verify project: %w", err)
		}
		if !projectExists {
			return ErrProjectNotFound
		}

		// 3. Calculate new balance based on transaction type
		balanceBefore := container.Balance
		var balanceAfter float64
		isIncome := req.Type == models.TransactionIncomeGrant || req.Type == models.TransactionIncomeDonation

		if isIncome {
			balanceAfter = balanceBefore + req.Amount
		} else {
			balanceAfter = balanceBefore - req.Amount
		}

		// 4. Check for deficit situation
		isDeficit := balanceAfter < 0
		deficitWarning := false

		if isDeficit && !req.ConfirmDeficit {
			// User needs to confirm deficit
			response.DeficitWarning = true
			response.RequiresApproval = true
			response.Message = fmt.Sprintf(
				"الرصيد الحالي: %.2f %s. المبلغ المطلوب: %.2f %s. سيصبح الرصيد: %.2f %s (عجز). هل تريد المتابعة؟",
				balanceBefore, container.Currency,
				req.Amount, req.Currency,
				balanceAfter, container.Currency,
			)
			return nil // Don't create transaction, just return warning
		}

		if isDeficit {
			deficitWarning = true
		}

		// 5. Set transaction date
		transactionDate := time.Now()
		if req.TransactionDate != nil {
			transactionDate = *req.TransactionDate
		}

		// 6. Calculate local amount
		amountLocal := req.Amount * req.ExchangeRate

		// 7. Insert transaction
		err = tx.QueryRowContext(ctx, `
			INSERT INTO transactions (
				type, status, amount, currency, exchange_rate, amount_local,
				container_id, project_id, description, description_ar,
				payee_payer, payee_payer_ar, reference_number,
				is_deficit, deficit_approved_by, deficit_warning_shown,
				balance_before, balance_after, transaction_date, created_by
			) VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
				$11, $12, $13, $14, $15, $16, $17, $18, $19, $20
			) RETURNING id, transaction_number, created_at
		`,
			req.Type, models.TransactionCompleted, req.Amount, req.Currency,
			req.ExchangeRate, amountLocal, req.ContainerID, req.ProjectID,
			req.Description, req.DescriptionAr, req.PayeePayer, req.PayeePayerAr,
			req.ReferenceNumber, isDeficit, userID, deficitWarning,
			balanceBefore, balanceAfter, transactionDate, userID,
		).Scan(&transaction.ID, &transaction.TransactionNumber, &transaction.CreatedAt)

		if err != nil {
			return fmt.Errorf("failed to insert transaction: %w", err)
		}

		// 8. Update container balance
		_, err = tx.ExecContext(ctx, `
			UPDATE financial_containers
			SET balance = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`, balanceAfter, req.ContainerID)

		if err != nil {
			return fmt.Errorf("failed to update container balance: %w", err)
		}

		// 9. Build complete transaction object
		transaction.Type = req.Type
		transaction.Status = models.TransactionCompleted
		transaction.Amount = req.Amount
		transaction.Currency = req.Currency
		transaction.ExchangeRate = req.ExchangeRate
		transaction.AmountLocal = amountLocal
		transaction.ContainerID = req.ContainerID
		transaction.ProjectID = req.ProjectID
		transaction.Description = req.Description
		transaction.DescriptionAr = req.DescriptionAr
		transaction.PayeePayer = req.PayeePayer
		transaction.PayeePayerAr = req.PayeePayerAr
		transaction.ReferenceNumber = req.ReferenceNumber
		transaction.IsDeficit = isDeficit
		transaction.DeficitWarningShown = deficitWarning
		transaction.BalanceBefore = &balanceBefore
		transaction.BalanceAfter = &balanceAfter
		transaction.TransactionDate = transactionDate
		transaction.CreatedBy = userID

		if isDeficit {
			transaction.DeficitApprovedBy = &userID
		}

		response.Transaction = transaction
		response.DeficitWarning = false
		response.RequiresApproval = false
		response.Message = "تمت العملية بنجاح"

		return nil
	})

	if err != nil {
		return nil, err
	}

	// If we got here without a transaction ID, it means we returned early with a warning
	if response.Transaction.ID == uuid.Nil && response.DeficitWarning {
		return &response, nil
	}

	return &response, nil
}

// GetTransactionByID retrieves a transaction by ID
func (s *TransactionService) GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	var t models.Transaction

	err := s.db.QueryRowContext(ctx, `
		SELECT id, transaction_number, type, status, amount, currency, exchange_rate, amount_local,
			   container_id, project_id, description, description_ar, payee_payer, payee_payer_ar,
			   reference_number, is_deficit, deficit_approved_by, deficit_warning_shown,
			   balance_before, balance_after, attachment_url, transaction_date,
			   created_at, updated_at, created_by, reversed_at, reversed_by, reversal_reason
		FROM transactions
		WHERE id = $1
	`, id).Scan(
		&t.ID, &t.TransactionNumber, &t.Type, &t.Status, &t.Amount, &t.Currency,
		&t.ExchangeRate, &t.AmountLocal, &t.ContainerID, &t.ProjectID,
		&t.Description, &t.DescriptionAr, &t.PayeePayer, &t.PayeePayerAr,
		&t.ReferenceNumber, &t.IsDeficit, &t.DeficitApprovedBy, &t.DeficitWarningShown,
		&t.BalanceBefore, &t.BalanceAfter, &t.AttachmentURL, &t.TransactionDate,
		&t.CreatedAt, &t.UpdatedAt, &t.CreatedBy, &t.ReversedAt, &t.ReversedBy, &t.ReversalReason,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("transaction not found")
	}
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// ReverseTransaction creates a reversal transaction
func (s *TransactionService) ReverseTransaction(ctx context.Context, transactionID uuid.UUID, reason string, userID uuid.UUID) error {
	return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Get original transaction
		var original models.Transaction
		err := tx.QueryRowContext(ctx, `
			SELECT id, type, amount, currency, exchange_rate, container_id, project_id, status
			FROM transactions
			WHERE id = $1
			FOR UPDATE
		`, transactionID).Scan(
			&original.ID, &original.Type, &original.Amount, &original.Currency,
			&original.ExchangeRate, &original.ContainerID, &original.ProjectID, &original.Status,
		)
		if err != nil {
			return err
		}

		if original.Status == models.TransactionReversed {
			return errors.New("transaction already reversed")
		}

		// Mark original as reversed
		_, err = tx.ExecContext(ctx, `
			UPDATE transactions
			SET status = $1, reversed_at = CURRENT_TIMESTAMP, reversed_by = $2, reversal_reason = $3
			WHERE id = $4
		`, models.TransactionReversed, userID, reason, transactionID)
		if err != nil {
			return err
		}

		// Create opposite transaction
		isIncome := original.Type == models.TransactionIncomeGrant || original.Type == models.TransactionIncomeDonation
		var reversalType models.TransactionType
		if isIncome {
			reversalType = models.TransactionDisbursementOperational
		} else {
			reversalType = models.TransactionIncomeDonation
		}

		_, err = tx.ExecContext(ctx, `
			INSERT INTO transactions (
				type, status, amount, currency, exchange_rate, amount_local,
				container_id, project_id, description, created_by, original_transaction_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`,
			reversalType, models.TransactionCompleted, original.Amount, original.Currency,
			original.ExchangeRate, original.Amount*original.ExchangeRate,
			original.ContainerID, original.ProjectID,
			"عكس قيد: "+reason, userID, transactionID,
		)

		return err
	})
}
