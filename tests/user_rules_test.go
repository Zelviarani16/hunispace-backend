package tests

import (
	"testing"

	"hunispace/app/model"
	"hunispace/app/service"
)

func TestValidateRegister(t *testing.T) {
	valid := model.RegisterRequest{
		FullName: "Budi Santoso", Email: "budi@example.com", Password: "rahasia123", PhoneNumber: "08123456789",
	}

	tests := []struct {
		name      string
		mutate    func(r *model.RegisterRequest)
		wantField string // kosong berarti data harus valid
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
			fields := service.ValidateRegister(req)

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

func TestValidateProfileUpdate(t *testing.T) {
	s := func(v string) *string { return &v }
	tests := []struct {
		name    string
		req     model.UpdateProfileRequest
		invalid bool
	}{
		{"nama saja", model.UpdateProfileRequest{FullName: s("Budi")}, false},
		{"telepon saja", model.UpdateProfileRequest{PhoneNumber: s("0812")}, false},
		{"telepon dikosongkan boleh", model.UpdateProfileRequest{PhoneNumber: s("")}, false},
		{"tidak ada field", model.UpdateProfileRequest{}, true},
		{"nama kosong", model.UpdateProfileRequest{FullName: s("  ")}, true},
		{"telepon terlalu panjang", model.UpdateProfileRequest{PhoneNumber: s("123456789012345678901")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ValidateProfileUpdate(tt.req)
			if (len(got) > 0) != tt.invalid {
				t.Fatalf("invalid=%v, fields=%v", tt.invalid, got)
			}
		})
	}
}

func TestValidateChangePassword(t *testing.T) {
	tests := []struct {
		name    string
		req     model.ChangePasswordRequest
		invalid bool
	}{
		{"valid", model.ChangePasswordRequest{OldPassword: "lama12345", NewPassword: "baru12345"}, false},
		{"lama kosong", model.ChangePasswordRequest{OldPassword: "", NewPassword: "baru12345"}, true},
		{"baru terlalu pendek", model.ChangePasswordRequest{OldPassword: "lama12345", NewPassword: "abc"}, true},
		{"baru sama dengan lama", model.ChangePasswordRequest{OldPassword: "samasama1", NewPassword: "samasama1"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.ValidateChangePassword(tt.req)
			if (len(got) > 0) != tt.invalid {
				t.Fatalf("invalid=%v, fields=%v", tt.invalid, got)
			}
		})
	}
}
