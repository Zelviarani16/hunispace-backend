package service

import (
	"fmt"
	"strings"
	"time"

	"hunispace/app/model"
)

// ValidBillingPeriod memeriksa format periode "YYYY-MM", misalnya "2026-11".
func ValidBillingPeriod(period string) bool {
	_, err := time.Parse("2006-01", period)
	return err == nil
}

// BuildInvoiceNumber menyusun nomor tagihan, misalnya INV-202611-12-A1B2C3.
// suffix acak mencegah nomor dapat ditebak.
func BuildInvoiceNumber(period string, leaseID int64, suffix string) string {
	return fmt.Sprintf("INV-%s-%d-%s", strings.ReplaceAll(period, "-", ""), leaseID, strings.ToUpper(suffix))
}

// ValidateInvoiceRequest memeriksa permintaan tagihan dan mengurai due_date.
func ValidateInvoiceRequest(req model.InvoiceCreateRequest) (due time.Time, fields map[string]string) {
	fields = map[string]string{}
	if req.LeaseID <= 0 {
		fields["lease_id"] = "wajib diisi"
	}
	if !ValidBillingPeriod(req.BillingPeriod) {
		fields["billing_period"] = "format harus YYYY-MM, mis. 2026-11"
	}
	if req.Amount <= 0 {
		fields["amount"] = "harus lebih dari 0"
	}
	due, err := time.Parse(dateLayout, req.DueDate)
	if err != nil {
		fields["due_date"] = "format harus YYYY-MM-DD"
	}
	return due, fields
}

// ValidateInvoicePay memeriksa konfirmasi pembayaran dari penghuni.
func ValidateInvoicePay(req model.InvoicePayRequest) map[string]string {
	fields := map[string]string{}
	if m := strings.TrimSpace(req.PaymentMethod); m == "" || len(m) > 30 {
		fields["payment_method"] = "wajib diisi, maksimal 30 karakter"
	}
	if strings.TrimSpace(req.ProofURL) == "" {
		fields["proof_url"] = "wajib diisi"
	}
	return fields
}

// ValidateInvoiceReject: alasan penolakan wajib diisi agar penghuni tahu apa yang harus diperbaiki.
func ValidateInvoiceReject(req model.InvoiceRejectRequest) map[string]string {
	fields := map[string]string{}
	if n := strings.TrimSpace(req.Note); n == "" || len(n) > 255 {
		fields["note"] = "wajib diisi, maksimal 255 karakter"
	}
	return fields
}
