package service

import (
	"fmt"
	"strings"
	"time"

	"hunispace/app/model"
)

func validBillingPeriod(period string) bool {
	_, err := time.Parse("2006-01", period)
	return err == nil
}

func buildInvoiceNumber(period string, leaseID int64, suffix string) string {
	return fmt.Sprintf("INV-%s-%d-%s", strings.ReplaceAll(period, "-", ""), leaseID, strings.ToUpper(suffix))
}

func validateInvoiceRequest(req model.InvoiceCreateRequest) (due time.Time, fields map[string]string) {
	fields = map[string]string{}
	if req.LeaseID <= 0 {
		fields["lease_id"] = "wajib diisi"
	}
	if !validBillingPeriod(req.BillingPeriod) {
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

func validateInvoicePay(req model.InvoicePayRequest) map[string]string {
	fields := map[string]string{}
	if m := strings.TrimSpace(req.PaymentMethod); m == "" || len(m) > 30 {
		fields["payment_method"] = "wajib diisi, maksimal 30 karakter"
	}
	if strings.TrimSpace(req.ProofURL) == "" {
		fields["proof_url"] = "wajib diisi"
	}
	return fields
}
