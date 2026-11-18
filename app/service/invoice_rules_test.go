package service

import (
	"testing"

	"hunispace/app/model"
)

func TestValidBillingPeriod(t *testing.T) {
	tests := []struct {
		period string
		want   bool
	}{
		{"2026-11", true},
		{"2026-01", true},
		{"2026-12", true},
		{"2026-13", false},
		{"2026-00", false},
		{"2026-1", false},
		{"26-11", false},
		{"", false},
		{"2026-11-01", false},
	}
	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			if got := validBillingPeriod(tt.period); got != tt.want {
				t.Fatalf("validBillingPeriod(%q) = %v, seharusnya %v", tt.period, got, tt.want)
			}
		})
	}
}

func TestBuildInvoiceNumber(t *testing.T) {
	got := buildInvoiceNumber("2026-11", 12, "a1b2c3")
	if want := "INV-202611-12-A1B2C3"; got != want {
		t.Fatalf("buildInvoiceNumber = %q, seharusnya %q", got, want)
	}
}

func TestValidateInvoiceRequest(t *testing.T) {
	valid := model.InvoiceCreateRequest{LeaseID: 1, BillingPeriod: "2026-11", Amount: 1500000, DueDate: "2026-11-10"}

	tests := []struct {
		name      string
		mutate    func(r *model.InvoiceCreateRequest)
		wantField string
	}{
		{"data valid", func(r *model.InvoiceCreateRequest) {}, ""},
		{"kontrak kosong", func(r *model.InvoiceCreateRequest) { r.LeaseID = 0 }, "lease_id"},
		{"periode salah", func(r *model.InvoiceCreateRequest) { r.BillingPeriod = "2026-13" }, "billing_period"},
		{"jumlah nol", func(r *model.InvoiceCreateRequest) { r.Amount = 0 }, "amount"},
		{"jatuh tempo salah format", func(r *model.InvoiceCreateRequest) { r.DueDate = "10/11/2026" }, "due_date"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			tt.mutate(&req)
			_, fields := validateInvoiceRequest(req)
			if tt.wantField == "" && len(fields) != 0 {
				t.Fatalf("seharusnya valid: %v", fields)
			}
			if tt.wantField != "" {
				if _, ok := fields[tt.wantField]; !ok {
					t.Fatalf("seharusnya ada error pada %q: %v", tt.wantField, fields)
				}
			}
		})
	}
}

func TestValidateInvoicePay(t *testing.T) {
	if f := validateInvoicePay(model.InvoicePayRequest{PaymentMethod: "transfer", ProofURL: "https://x/bukti.jpg"}); len(f) != 0 {
		t.Fatalf("seharusnya valid: %v", f)
	}
	if f := validateInvoicePay(model.InvoicePayRequest{PaymentMethod: " ", ProofURL: ""}); len(f) != 2 {
		t.Fatalf("seharusnya dua error: %v", f)
	}
}
