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
	CreatedAt     time.Time  `json:"created_at"`
}

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
