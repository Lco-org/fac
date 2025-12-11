package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type ReceiptService struct {
	db *database.DB
}

func NewReceiptService(db *database.DB) *ReceiptService {
	return &ReceiptService{db: db}
}

type ReceiptData struct {
	Transaction        models.Transaction         `json:"transaction"`
	Container          models.FinancialContainer  `json:"container"`
	Project            models.Project             `json:"project"`
	CreatedByUser      UserBasicInfo              `json:"created_by_user"`
	RelatedTransaction *models.Transaction        `json:"related_transaction,omitempty"`
	RelatedContainer   *models.FinancialContainer `json:"related_container,omitempty"`
	Employee           *models.Employee           `json:"employee,omitempty"`
	Member             *models.Member             `json:"member,omitempty"`
}

type UserBasicInfo struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	FullNameAr *string   `json:"full_name_ar,omitempty"`
}

func (s *ReceiptService) GetReceiptData(ctx context.Context, transactionID uuid.UUID) (*ReceiptData, error) {
	var receipt ReceiptData

	err := s.db.QueryRowContext(ctx, `
		SELECT 
			t.id, t.transaction_number, t.type, t.status, t.amount, t.currency, 
			t.exchange_rate, t.amount_local, t.container_id, 
			COALESCE(t.project_id, '00000000-0000-0000-0000-000000000000'),
			t.employee_id, t.member_id,
			t.description, t.description_ar, t.payee_payer, t.payee_payer_ar,
			t.reference_number, t.is_deficit, t.deficit_approved_by, t.deficit_warning_shown,
			t.balance_before, t.balance_after, t.attachment_url, t.transaction_date,
			t.created_at, t.updated_at, t.created_by, t.reversed_at, t.reversed_by, t.reversal_reason,
			t.original_transaction_id
		FROM transactions t
		WHERE t.id = $1
	`, transactionID).Scan(
		&receipt.Transaction.ID,
		&receipt.Transaction.TransactionNumber,
		&receipt.Transaction.Type,
		&receipt.Transaction.Status,
		&receipt.Transaction.Amount,
		&receipt.Transaction.Currency,
		&receipt.Transaction.ExchangeRate,
		&receipt.Transaction.AmountLocal,
		&receipt.Transaction.ContainerID,
		&receipt.Transaction.ProjectID,
		&receipt.Transaction.EmployeeID,
		&receipt.Transaction.MemberID,
		&receipt.Transaction.Description,
		&receipt.Transaction.DescriptionAr,
		&receipt.Transaction.PayeePayer,
		&receipt.Transaction.PayeePayerAr,
		&receipt.Transaction.ReferenceNumber,
		&receipt.Transaction.IsDeficit,
		&receipt.Transaction.DeficitApprovedBy,
		&receipt.Transaction.DeficitWarningShown,
		&receipt.Transaction.BalanceBefore,
		&receipt.Transaction.BalanceAfter,
		&receipt.Transaction.AttachmentURL,
		&receipt.Transaction.TransactionDate,
		&receipt.Transaction.CreatedAt,
		&receipt.Transaction.UpdatedAt,
		&receipt.Transaction.CreatedBy,
		&receipt.Transaction.ReversedAt,
		&receipt.Transaction.ReversedBy,
		&receipt.Transaction.ReversalReason,
		&receipt.Transaction.OriginalTransactionID,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("transaction not found")
	}
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT 
			id, name, name_ar, type, currency, balance,
			bank_name, bank_name_ar, account_number, iban, swift_code,
			location, location_ar, is_active, created_at, updated_at, created_by
		FROM financial_containers
		WHERE id = $1
	`, receipt.Transaction.ContainerID).Scan(
		&receipt.Container.ID,
		&receipt.Container.Name,
		&receipt.Container.NameAr,
		&receipt.Container.Type,
		&receipt.Container.Currency,
		&receipt.Container.Balance,
		&receipt.Container.BankName,
		&receipt.Container.BankNameAr,
		&receipt.Container.AccountNumber,
		&receipt.Container.IBAN,
		&receipt.Container.SwiftCode,
		&receipt.Container.Location,
		&receipt.Container.LocationAr,
		&receipt.Container.IsActive,
		&receipt.Container.CreatedAt,
		&receipt.Container.UpdatedAt,
		&receipt.Container.CreatedBy,
	)

	if err != nil {
		return nil, err
	}

	if receipt.Transaction.ProjectID != uuid.Nil {
		err = s.db.QueryRowContext(ctx, `
			SELECT 
				id, code, name, name_ar, description,
				donor_name, donor_name_ar, start_date, end_date,
				budget, currency, is_active, created_at, updated_at, created_by
			FROM projects
			WHERE id = $1
		`, receipt.Transaction.ProjectID).Scan(
			&receipt.Project.ID,
			&receipt.Project.Code,
			&receipt.Project.Name,
			&receipt.Project.NameAr,
			&receipt.Project.Description,
			&receipt.Project.DonorName,
			&receipt.Project.DonorNameAr,
			&receipt.Project.StartDate,
			&receipt.Project.EndDate,
			&receipt.Project.Budget,
			&receipt.Project.Currency,
			&receipt.Project.IsActive,
			&receipt.Project.CreatedAt,
			&receipt.Project.UpdatedAt,
			&receipt.Project.CreatedBy,
		)

		if err != nil {
			return nil, err
		}
	} else {
		receipt.Project.ID = uuid.Nil
		receipt.Project.Code = "-"
		receipt.Project.Name = "تحويل داخلي"
		nameAr := "تحويل داخلي"
		receipt.Project.NameAr = &nameAr
		receipt.Project.Currency = receipt.Transaction.Currency
		receipt.Project.IsActive = true
	}

	if receipt.Transaction.EmployeeID != nil {
		var employee models.Employee
		err = s.db.QueryRowContext(ctx, `
			SELECT id, employee_number, full_name, full_name_ar, position, position_ar
			FROM employees WHERE id = $1
		`, receipt.Transaction.EmployeeID).Scan(
			&employee.ID, &employee.EmployeeNumber, &employee.FullName, &employee.FullNameAr, &employee.Position, &employee.PositionAr,
		)
		if err == nil {
			receipt.Employee = &employee
		}
	}

	if receipt.Transaction.MemberID != nil {
		var member models.Member
		err = s.db.QueryRowContext(ctx, `
			SELECT id, membership_number, full_name, full_name_ar, member_type
			FROM members WHERE id = $1
		`, receipt.Transaction.MemberID).Scan(
			&member.ID, &member.MembershipNumber, &member.FullName, &member.FullNameAr, &member.MemberType,
		)
		if err == nil {
			receipt.Member = &member
		}
	}

	err = s.db.QueryRowContext(ctx, `
		SELECT id, full_name, full_name_ar
		FROM users
		WHERE id = $1
	`, receipt.Transaction.CreatedBy).Scan(
		&receipt.CreatedByUser.ID,
		&receipt.CreatedByUser.FullName,
		&receipt.CreatedByUser.FullNameAr,
	)

	if err != nil {
		return nil, err
	}

	if receipt.Transaction.Type == models.TransactionTransfer {
		var relatedTxID uuid.UUID

		if receipt.Transaction.OriginalTransactionID != nil {
			relatedTxID = *receipt.Transaction.OriginalTransactionID
		} else {
			err := s.db.QueryRowContext(ctx, `
				SELECT id FROM transactions WHERE original_transaction_id = $1
			`, receipt.Transaction.ID).Scan(&relatedTxID)

			if err != nil && err != sql.ErrNoRows {
				fmt.Printf("Error finding related transfer transaction: %v\n", err)
			}
		}

		if relatedTxID != uuid.Nil {
			var relatedTx models.Transaction
			err := s.db.QueryRowContext(ctx, `
				SELECT id, container_id, type, amount, currency
				FROM transactions WHERE id = $1
			`, relatedTxID).Scan(
				&relatedTx.ID,
				&relatedTx.ContainerID,
				&relatedTx.Type,
				&relatedTx.Amount,
				&relatedTx.Currency,
			)

			if err == nil {
				receipt.RelatedTransaction = &relatedTx

				var relatedContainer models.FinancialContainer
				err = s.db.QueryRowContext(ctx, `
					SELECT id, name, name_ar, type, currency
					FROM financial_containers WHERE id = $1
				`, relatedTx.ContainerID).Scan(
					&relatedContainer.ID,
					&relatedContainer.Name,
					&relatedContainer.NameAr,
					&relatedContainer.Type,
					&relatedContainer.Currency,
				)

				if err == nil {
					receipt.RelatedContainer = &relatedContainer
				}
			}
		}
	}

	return &receipt, nil
}
