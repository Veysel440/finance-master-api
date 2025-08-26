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
	AuthFailed         = E("auth_failed", 401, "authentication failed")
	Unauthorized       = E("unauthorized", 401, "unauthorized")
	Forbidden          = E("forbidden", 403, "forbidden")
	NotFound           = E("not_found", 404, "not found")
	Conflict           = E("conflict", 409, "conflict")
	TOTPRequired       = E("totp_required", 401, "totp required")
	TOTPInvalid        = E("totp_invalid", 401, "invalid totp")
	InvalidRefresh     = E("invalid_refresh", 401, "invalid refresh token")
	ValidationFailed   = func(msg string) *AppError { return E("validation_failed", 400, msg) }
	Internal           = E("internal", 500, "internal server error")

	HasTransactions   = E("has_transactions", 409, "resource has linked transactions")
	CaptchaRequired   = E("captcha_required", 401, "captcha required")
	SlowDown          = E("slow_down", 429, "too many attempts, slow down")
	InsecureTransport = E("insecure_transport", 426, "https required")
)

type RetryAfterError struct {
	*AppError
	Seconds int
}

func (e *RetryAfterError) Error() string { return e.AppError.Error() }
func SlowDownAfter(seconds int) error {
	if seconds < 1 {
		seconds = 1
	}
	return &RetryAfterError{AppError: SlowDown, Seconds: seconds}
}
