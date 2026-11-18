package service

import (
	"testing"

	"hunispace/app/model"
)

func TestValidateLeaseRequest(t *testing.T) {
	valid := model.LeaseCreateRequest{TenantID: 2, RoomID: 3, StartDate: "2026-12-01", EndDate: "2027-05-31"}

	tests := []struct {
		name      string
		mutate    func(r *model.LeaseCreateRequest)
		wantField string
	}{
		{"data valid", func(r *model.LeaseCreateRequest) {}, ""},
		{"tenant kosong", func(r *model.LeaseCreateRequest) { r.TenantID = 0 }, "tenant_id"},
		{"kamar kosong", func(r *model.LeaseCreateRequest) { r.RoomID = 0 }, "room_id"},
		{"format tanggal mulai salah", func(r *model.LeaseCreateRequest) { r.StartDate = "01-12-2026" }, "start_date"},
		{"selesai sama dengan mulai", func(r *model.LeaseCreateRequest) { r.EndDate = "2026-12-01" }, "end_date"},
		{"selesai sebelum mulai", func(r *model.LeaseCreateRequest) { r.EndDate = "2026-11-30" }, "end_date"},
		{"selesai sehari setelah mulai", func(r *model.LeaseCreateRequest) { r.EndDate = "2026-12-02" }, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			tt.mutate(&req)
			_, _, fields := validateLeaseRequest(req)
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
