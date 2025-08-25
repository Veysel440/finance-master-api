package http

import (
	"net/http"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/validation"
)

func BindAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := DecodeStrict(r, dst); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return false
	}
	if err := validation.ValidateStruct(dst); err != nil {
		WriteAppError(w, errs.ValidationFailed(validation.ValidationMessage(err)))
		return false
	}
	return true
}
