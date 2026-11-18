package service

import (
	"testing"

	"hunispace/app/model"
)

func TestValidateTicketCreate(t *testing.T) {
	t.Run("prioritas kosong menjadi medium", func(t *testing.T) {
		req := model.TicketCreateRequest{Title: "AC mati", Description: "AC tidak dingin sejak kemarin"}
		if fields := validateTicketCreate(&req); len(fields) != 0 {
			t.Fatalf("seharusnya valid: %v", fields)
		}
		if req.Priority != model.PriorityMedium {
			t.Fatalf("prioritas = %q, seharusnya medium", req.Priority)
		}
	})

	tests := []struct {
		name      string
		req       model.TicketCreateRequest
		wantField string
	}{
		{"judul kosong", model.TicketCreateRequest{Description: "x", Priority: "low"}, "title"},
		{"deskripsi kosong", model.TicketCreateRequest{Title: "x", Priority: "low"}, "description"},
		{"prioritas tidak dikenal", model.TicketCreateRequest{Title: "x", Description: "y", Priority: "urgent"}, "priority"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.req
			fields := validateTicketCreate(&req)
			if _, ok := fields[tt.wantField]; !ok {
				t.Fatalf("seharusnya ada error pada %q: %v", tt.wantField, fields)
			}
		})
	}
}

func TestValidateTicketStatus(t *testing.T) {
	tests := []struct {
		name      string
		req       model.TicketStatusRequest
		wantField string
	}{
		{"valid", model.TicketStatusRequest{Status: "in_progress", Version: 1}, ""},
		{"status tidak dikenal", model.TicketStatusRequest{Status: "done", Version: 1}, "status"},
		{"version nol", model.TicketStatusRequest{Status: "resolved", Version: 0}, "version"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := validateTicketStatus(tt.req)
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
