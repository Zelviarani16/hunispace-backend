package service

import (
	"net/mail"
	"strings"

	"hunispace/app/model"
)

func validateRegister(req model.RegisterRequest) map[string]string {
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
