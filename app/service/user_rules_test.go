package service

import (
	"testing"

	"hunispace/app/model"
)

func TestValidateRegister(t *testing.T) {
	valid := model.RegisterRequest{
		FullName: "Budi Santoso", Email: "budi@example.com", Password: "rahasia123", PhoneNumber: "08123456789",
	}

	tests := []struct {
		name      string
		mutate    func(r *model.RegisterRequest)
		wantField string
	}{
		{"data valid", func(r *model.RegisterRequest) {}, ""},
		{"nama kosong", func(r *model.RegisterRequest) { r.FullName = "   " }, "full_name"},
		{"email salah format", func(r *model.RegisterRequest) { r.Email = "bukan-email" }, "email"},
		{"password tepat 8 karakter", func(r *model.RegisterRequest) { r.Password = "12345678" }, ""},
		{"password 7 karakter", func(r *model.RegisterRequest) { r.Password = "1234567" }, "password"},
		{"telepon terlalu panjang", func(r *model.RegisterRequest) { r.PhoneNumber = "0812345678901234567890" }, "phone_number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid
			tt.mutate(&req)
			fields := validateRegister(req)

			if tt.wantField == "" {
				if len(fields) != 0 {
					t.Fatalf("seharusnya valid, tetapi ada error: %v", fields)
				}
				return
			}
			if _, ok := fields[tt.wantField]; !ok {
				t.Fatalf("seharusnya ada error pada %q, hasil: %v", tt.wantField, fields)
			}
		})
	}
}
