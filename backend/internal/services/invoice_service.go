package services

import (
	"bytes"
	"context"
	"cso-erp/internal/database"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// InvoiceService handles invoice generation
type InvoiceService struct {
	db *database.DB
}

// NewInvoiceService creates a new invoice service
func NewInvoiceService(db *database.DB) *InvoiceService {
	return &InvoiceService{db: db}
}

// InvoiceData represents data for generating an invoice
type InvoiceData struct {
	ID              uuid.UUID `json:"id"`
	InvoiceNumber   string    `json:"invoice_number"`
	TransactionID   uuid.UUID `json:"transaction_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Description     string    `json:"description"`

	// Payer/Payee Info
	PayerName    string `json:"payer_name"`
	PayerAddress string `json:"payer_address,omitempty"`
	PayeeName    string `json:"payee_name"`
	PayeeAddress string `json:"payee_address,omitempty"`

	// Organization Info
	OrgName    string `json:"org_name"`
	OrgAddress string `json:"org_address,omitempty"`
	OrgPhone   string `json:"org_phone,omitempty"`
	OrgEmail   string `json:"org_email,omitempty"`
	OrgLogo    string `json:"org_logo,omitempty"`

	// Dates
	IssueDate time.Time  `json:"issue_date"`
	DueDate   *time.Time `json:"due_date,omitempty"`

	// Items (for detailed invoices)
	Items []InvoiceItem `json:"items,omitempty"`

	// Totals
	Subtotal float64 `json:"subtotal"`
	Tax      float64 `json:"tax"`
	Total    float64 `json:"total"`

	// Status
	Status string `json:"status"` // draft, issued, paid, cancelled
	Notes  string `json:"notes,omitempty"`

	CreatedAt time.Time  `json:"created_at"`
	CreatedBy *uuid.UUID `json:"created_by"`
}

// InvoiceItem represents a line item in an invoice
type InvoiceItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}

// GenerateInvoiceNumber generates a unique invoice number
func (s *InvoiceService) GenerateInvoiceNumber(ctx context.Context) (string, error) {
	year := time.Now().Year()

	// Get count of invoices this year
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM invoices 
		WHERE EXTRACT(YEAR FROM created_at) = $1
	`, year).Scan(&count)

	if err != nil {
		count = 0
	}

	return fmt.Sprintf("INV-%d-%05d", year, count+1), nil
}

// GenerateInvoiceFromTransaction creates an invoice from a transaction
func (s *InvoiceService) GenerateInvoiceFromTransaction(ctx context.Context, transactionID uuid.UUID, createdBy *uuid.UUID) (*InvoiceData, error) {
	// Get transaction details
	var invoice InvoiceData
	err := s.db.QueryRowContext(ctx, `
		SELECT 
			t.id, t.type, t.amount, t.currency, t.description,
			COALESCE(t.payer_name, '') as payer_name,
			COALESCE(t.payee_name, '') as payee_name,
			t.created_at
		FROM transactions t
		WHERE t.id = $1
	`, transactionID).Scan(
		&invoice.TransactionID, &invoice.TransactionType, &invoice.Amount,
		&invoice.Currency, &invoice.Description, &invoice.PayerName,
		&invoice.PayeeName, &invoice.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("transaction not found: %w", err)
	}

	// Generate invoice number
	invoiceNum, err := s.GenerateInvoiceNumber(ctx)
	if err != nil {
		return nil, err
	}

	invoice.ID = uuid.New()
	invoice.InvoiceNumber = invoiceNum
	invoice.IssueDate = time.Now()
	invoice.Subtotal = invoice.Amount
	invoice.Tax = 0
	invoice.Total = invoice.Amount
	invoice.Status = "issued"
	invoice.CreatedBy = createdBy

	// Save to database
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO invoices (id, invoice_number, transaction_id, amount, currency, 
			payer_name, payee_name, description, issue_date, total, status, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, invoice.ID, invoice.InvoiceNumber, invoice.TransactionID, invoice.Amount,
		invoice.Currency, invoice.PayerName, invoice.PayeeName, invoice.Description,
		invoice.IssueDate, invoice.Total, invoice.Status, invoice.CreatedBy, time.Now())

	if err != nil {
		return nil, fmt.Errorf("failed to save invoice: %w", err)
	}

	return &invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (s *InvoiceService) GetInvoice(ctx context.Context, invoiceID uuid.UUID) (*InvoiceData, error) {
	var invoice InvoiceData
	err := s.db.QueryRowContext(ctx, `
		SELECT id, invoice_number, transaction_id, amount, currency, 
			payer_name, payee_name, description, issue_date, total, status, 
			created_by, created_at
		FROM invoices
		WHERE id = $1
	`, invoiceID).Scan(
		&invoice.ID, &invoice.InvoiceNumber, &invoice.TransactionID, &invoice.Amount,
		&invoice.Currency, &invoice.PayerName, &invoice.PayeeName, &invoice.Description,
		&invoice.IssueDate, &invoice.Total, &invoice.Status, &invoice.CreatedBy, &invoice.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &invoice, nil
}

// ListInvoices lists all invoices with pagination
func (s *InvoiceService) ListInvoices(ctx context.Context, limit, offset int) ([]InvoiceData, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, invoice_number, transaction_id, amount, currency, 
			payer_name, payee_name, description, issue_date, total, status, 
			created_by, created_at
		FROM invoices
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []InvoiceData
	for rows.Next() {
		var inv InvoiceData
		if err := rows.Scan(
			&inv.ID, &inv.InvoiceNumber, &inv.TransactionID, &inv.Amount,
			&inv.Currency, &inv.PayerName, &inv.PayeeName, &inv.Description,
			&inv.IssueDate, &inv.Total, &inv.Status, &inv.CreatedBy, &inv.CreatedAt,
		); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}

	return invoices, nil
}

// GenerateInvoiceHTML generates HTML version of invoice for printing/PDF
func (s *InvoiceService) GenerateInvoiceHTML(invoice *InvoiceData, orgName, orgAddress string) string {
	var buf bytes.Buffer

	buf.WriteString(`<!DOCTYPE html>
<html dir="rtl" lang="ar">
<head>
    <meta charset="UTF-8">
    <title>فاتورة - ` + invoice.InvoiceNumber + `</title>
    <style>
        body { font-family: 'Arial', sans-serif; margin: 40px; direction: rtl; }
        .header { display: flex; justify-content: space-between; margin-bottom: 40px; }
        .logo { font-size: 24px; font-weight: bold; color: #2c3e50; }
        .invoice-info { text-align: left; }
        .invoice-number { font-size: 20px; font-weight: bold; }
        .parties { display: flex; justify-content: space-between; margin-bottom: 40px; }
        .party { width: 45%; }
        .party h3 { color: #2c3e50; border-bottom: 2px solid #3498db; padding-bottom: 5px; }
        .details { margin-bottom: 40px; }
        .details table { width: 100%; border-collapse: collapse; }
        .details th, .details td { border: 1px solid #ddd; padding: 12px; text-align: right; }
        .details th { background-color: #3498db; color: white; }
        .totals { text-align: left; margin-top: 20px; }
        .totals table { margin-right: 0; margin-left: auto; }
        .totals td { padding: 8px 20px; }
        .total-row { font-size: 18px; font-weight: bold; background: #ecf0f1; }
        .footer { margin-top: 60px; text-align: center; color: #7f8c8d; font-size: 12px; }
        .status { padding: 5px 15px; border-radius: 20px; color: white; font-weight: bold; }
        .status.issued { background: #27ae60; }
        .status.paid { background: #3498db; }
        .status.cancelled { background: #e74c3c; }
    </style>
</head>
<body>
    <div class="header">
        <div class="logo">` + orgName + `</div>
        <div class="invoice-info">
            <div class="invoice-number">فاتورة رقم: ` + invoice.InvoiceNumber + `</div>
            <div>تاريخ الإصدار: ` + invoice.IssueDate.Format("2006-01-02") + `</div>
            <div>الحالة: <span class="status ` + invoice.Status + `">` + invoice.Status + `</span></div>
        </div>
    </div>

    <div class="parties">
        <div class="party">
            <h3>من:</h3>
            <p><strong>` + orgName + `</strong></p>
            <p>` + orgAddress + `</p>
        </div>
        <div class="party">
            <h3>إلى:</h3>
            <p><strong>` + invoice.PayerName + `</strong></p>
        </div>
    </div>

    <div class="details">
        <table>
            <thead>
                <tr>
                    <th>الوصف</th>
                    <th>المبلغ</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td>` + invoice.Description + `</td>
                    <td>` + fmt.Sprintf("%.2f", invoice.Amount) + ` ` + invoice.Currency + `</td>
                </tr>
            </tbody>
        </table>
    </div>

    <div class="totals">
        <table>
            <tr>
                <td>المجموع الفرعي:</td>
                <td>` + fmt.Sprintf("%.2f", invoice.Subtotal) + ` ` + invoice.Currency + `</td>
            </tr>
            <tr>
                <td>الضريبة:</td>
                <td>` + fmt.Sprintf("%.2f", invoice.Tax) + ` ` + invoice.Currency + `</td>
            </tr>
            <tr class="total-row">
                <td>المجموع الكلي:</td>
                <td>` + fmt.Sprintf("%.2f", invoice.Total) + ` ` + invoice.Currency + `</td>
            </tr>
        </table>
    </div>

    <div class="footer">
        <p>شكراً لتعاملكم معنا</p>
        <p>` + orgName + ` - نظام CSO-ERP</p>
    </div>
</body>
</html>`)

	return buf.String()
}
