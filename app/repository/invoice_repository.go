package repository

import (
	"context"

	"hunispace/app/model"
)

type InvoiceRepository interface {
	Create(ctx context.Context, db DBTX, inv model.Invoice) (model.Invoice, error)
	FindByID(ctx context.Context, db DBTX, id int64) (model.Invoice, error)
	FindAll(ctx context.Context, db DBTX, status string) ([]model.Invoice, error)
	FindByTenant(ctx context.Context, db DBTX, tenantID int64) ([]model.Invoice, error)
	MarkPendingVerification(ctx context.Context, db DBTX, id, tenantID int64, method, proofURL string) (bool, error)
	MarkPaid(ctx context.Context, db DBTX, id int64) (bool, error)
}

type invoicePostgresRepository struct{}

func NewInvoiceRepository() InvoiceRepository { return &invoicePostgresRepository{} }

const invoiceColumns = "id, lease_id, invoice_number, billing_period, amount, due_date, payment_method, proof_url, status, paid_at, created_at"

const invoiceColumnsAliased = "i.id, i.lease_id, i.invoice_number, i.billing_period, i.amount, i.due_date, i.payment_method, i.proof_url, i.status, i.paid_at, i.created_at"

func scanInvoice(row rowScanner) (model.Invoice, error) {
	var inv model.Invoice
	err := row.Scan(&inv.ID, &inv.LeaseID, &inv.InvoiceNumber, &inv.BillingPeriod, &inv.Amount,
		&inv.DueDate, &inv.PaymentMethod, &inv.ProofURL, &inv.Status, &inv.PaidAt, &inv.CreatedAt)
	return inv, err
}

func (r *invoicePostgresRepository) Create(ctx context.Context, db DBTX, inv model.Invoice) (model.Invoice, error) {
	created, err := scanInvoice(db.QueryRow(ctx,
		`INSERT INTO invoices (lease_id, invoice_number, billing_period, amount, due_date)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+invoiceColumns,
		inv.LeaseID, inv.InvoiceNumber, inv.BillingPeriod, inv.Amount, inv.DueDate))
	if err != nil {
		if isUniqueViolation(err) {
			return model.Invoice{}, ErrDuplicate
		}
		return model.Invoice{}, err
	}
	return created, nil
}

func (r *invoicePostgresRepository) FindByID(ctx context.Context, db DBTX, id int64) (model.Invoice, error) {
	inv, err := scanInvoice(db.QueryRow(ctx, `SELECT `+invoiceColumns+` FROM invoices WHERE id = $1`, id))
	if err != nil {
		return model.Invoice{}, mapNotFound(err)
	}
	return inv, nil
}

func (r *invoicePostgresRepository) FindAll(ctx context.Context, db DBTX, status string) ([]model.Invoice, error) {
	query := `SELECT ` + invoiceColumns + ` FROM invoices`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, status)
	}
	query += ` ORDER BY due_date DESC, id DESC`

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanInvoice)
}

func (r *invoicePostgresRepository) FindByTenant(ctx context.Context, db DBTX, tenantID int64) ([]model.Invoice, error) {
	rows, err := db.Query(ctx,
		`SELECT `+invoiceColumnsAliased+`
		 FROM invoices i
		 JOIN leases l ON l.id = i.lease_id
		 WHERE l.tenant_id = $1
		 ORDER BY i.due_date DESC, i.id DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	return collect(rows, scanInvoice)
}

func (r *invoicePostgresRepository) MarkPendingVerification(
	ctx context.Context, db DBTX, id, tenantID int64, method, proofURL string,
) (bool, error) {
	tag, err := db.Exec(ctx,
		`UPDATE invoices AS i
		 SET status = 'pending_verification', payment_method = $3, proof_url = $4
		 FROM leases AS l
		 WHERE i.id = $1
		   AND l.id = i.lease_id
		   AND l.tenant_id = $2
		   AND i.status IN ('unpaid', 'overdue')`,
		id, tenantID, method, proofURL)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

func (r *invoicePostgresRepository) MarkPaid(ctx context.Context, db DBTX, id int64) (bool, error) {
	tag, err := db.Exec(ctx,
		`UPDATE invoices SET status = 'paid', paid_at = now()
		 WHERE id = $1 AND status = 'pending_verification'`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}
