package model

import "time"

const (
	InvoiceUnpaid              = "unpaid"
	InvoicePendingVerification = "pending_verification"
	InvoicePaid                = "paid"
	InvoiceOverdue             = "overdue"
)

type Invoice struct {
	ID            int64      `json:"id"`
	LeaseID       int64      `json:"lease_id"`
	InvoiceNumber string     `json:"invoice_number"`
	BillingPeriod string     `json:"billing_period"`
	Amount        float64    `json:"amount"`
	DueDate       time.Time  `json:"due_date"`
	PaymentMethod *string    `json:"payment_method"`
	ProofURL      *string    `json:"proof_url"`
	Status        string     `json:"status"`
	PaidAt        *time.Time `json:"paid_at"`
	RejectionNote *string    `json:"rejection_note"`
	CreatedAt     time.Time  `json:"created_at"`
}

// InvoiceCreateRequest: billing_period berbentuk "2026-11", due_date "YYYY-MM-DD".
type InvoiceCreateRequest struct {
	LeaseID       int64   `json:"lease_id"`
	BillingPeriod string  `json:"billing_period"`
	Amount        float64 `json:"amount"`
	DueDate       string  `json:"due_date"`
}

type InvoicePayRequest struct {
	PaymentMethod string `json:"payment_method"`
	ProofURL      string `json:"proof_url"`
}

// InvoiceRejectRequest berisi alasan owner menolak bukti pembayaran.
type InvoiceRejectRequest struct {
	Note string `json:"note"`
}

// OverdueResult adalah isi data pada POST /invoices/mark-overdue.
type OverdueResult struct {
	Updated int64 `json:"updated" example:"3"`
}
