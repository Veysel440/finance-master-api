package errs

import "fmt"

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	HTTP    int    `json:"-"`
}

func (e *AppError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }
func E(code string, http int, msg string) *AppError {
	return &AppError{Code: code, HTTP: http, Message: msg}
}

var (
	InvalidCredentials = E("invalid_credentials", 401, "invalid credentials")
	Unauthorized       = E("unauthorized", 401, "unauthorized")
	Forbidden          = E("forbidden", 403, "forbidden")
	NotFound           = E("not_found", 404, "not found")
	Conflict           = E("conflict", 409, "conflict")
	TOTPRequired       = E("totp_required", 401, "totp required")
	TOTPInvalid        = E("totp_invalid", 401, "invalid totp")
	InvalidRefresh     = E("invalid_refresh", 401, "invalid refresh token")
	ValidationFailed   = func(msg string) *AppError { return E("validation_failed", 400, msg) }
	Internal           = E("internal", 500, "internal server error")
)
