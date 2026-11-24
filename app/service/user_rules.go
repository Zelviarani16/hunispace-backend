package service

import (
	"net/mail"
	"strings"

	"hunispace/app/model"
)

func ValidateRegister(req model.RegisterRequest) map[string]string {
	fields := map[string]string{}

	name := strings.TrimSpace(req.FullName)
	if name == "" || len(name) > 100 {
		fields["full_name"] = "wajib diisi, maksimal 100 karakter"
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(req.Email)); err != nil {
		fields["email"] = "format email tidak valid"
	}
	if len(req.Password) < 8 {
		fields["password"] = "minimal 8 karakter"
	}
	if len(req.PhoneNumber) > 20 {
		fields["phone_number"] = "maksimal 20 karakter"
	}
	return fields
}

func ValidateProfileUpdate(req model.UpdateProfileRequest) map[string]string {
	fields := map[string]string{}
	if req.FullName == nil && req.PhoneNumber == nil {
		fields["body"] = "kirim minimal satu field: full_name atau phone_number"
	}
	if req.FullName != nil {
		if n := strings.TrimSpace(*req.FullName); n == "" || len(n) > 100 {
			fields["full_name"] = "tidak boleh kosong, maksimal 100 karakter"
		}
	}
	if req.PhoneNumber != nil && len(*req.PhoneNumber) > 20 {
		fields["phone_number"] = "maksimal 20 karakter"
	}
	return fields
}

func ValidateChangePassword(req model.ChangePasswordRequest) map[string]string {
	fields := map[string]string{}
	if req.OldPassword == "" {
		fields["old_password"] = "wajib diisi"
	}
	if len(req.NewPassword) < 8 {
		fields["new_password"] = "minimal 8 karakter"
	} else if req.NewPassword == req.OldPassword {
		fields["new_password"] = "harus berbeda dari password lama"
	}
	return fields
}
