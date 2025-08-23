package http

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Veysel440/finance-master-api/internal/errs"
	mysqlerr "github.com/go-sql-driver/mysql"
)

type jsonError struct {
	Error   bool   `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func WriteJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteAppError(w http.ResponseWriter, e *errs.AppError) {
	WriteJSON(w, e.HTTP, jsonError{Error: true, Code: e.Code, Message: e.Message})
}

func Fail(w http.ResponseWriter, httpCode int, code, msg string) {
	WriteAppError(w, errs.E(code, httpCode, msg))
}

func FromError(w http.ResponseWriter, err error) {
	app := MapError(err)
	WriteAppError(w, app)
}

func MapError(err error) *errs.AppError {
	if err == nil {
		return errs.Internal
	}
	var app *errs.AppError
	if errors.As(err, &app) {
		return app
	}
	if errors.Is(err, sql.ErrNoRows) {
		return errs.NotFound
	}

	var me *mysqlerr.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return errs.Conflict
	}
	return errs.Internal
}
