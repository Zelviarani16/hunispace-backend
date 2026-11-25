package model

type WebResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type ErrorResponse struct {
	Success bool      `json:"success" example:"false"`
	Message string    `json:"message" example:"data tidak ditemukan"`
	Error   ErrorBody `json:"error"`
}

type AuthUser struct {
	UserID int64
	Role   string
}
