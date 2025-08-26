package http

import (
	"net/http"
	"strconv"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/security"
	"github.com/Veysel440/finance-master-api/internal/services"
	"github.com/Veysel440/finance-master-api/internal/validation"
	"github.com/go-chi/chi/v5"
)

type AuthHandlers struct{ S *services.AuthService }

type reqRegister struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email,max=320"`
	Password string `json:"password" validate:"required,min=10,max=256"`
}

type reqLogin struct {
	Email      string `json:"email"      validate:"required,email,max=320"`
	Password   string `json:"password"   validate:"required,min=6,max=256"`
	Totp       string `json:"totp,omitempty"        validate:"omitempty,len=6,numeric"`
	DeviceID   string `json:"deviceId,omitempty"    validate:"omitempty,max=128"`
	DeviceName string `json:"deviceName,omitempty"  validate:"omitempty,max=128"`
	Captcha    string `json:"captcha,omitempty"     validate:"omitempty,max=2048"`
}

type reqRefresh struct {
	Refresh string `json:"refresh" validate:"required,min=10"`
}

type reqTotpConfirm struct {
	Code string `json:"code" validate:"required,len=6,numeric"`
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var in reqRegister
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if err := validation.ValidateStruct(in); err != nil {
		WriteAppError(w, errs.ValidationFailed(validation.ValidationMessage(err)))
		return
	}
	if err := validation.ValidatePassword(in.Password, in.Email); err != nil {
		WriteAppError(w, errs.ValidationFailed("weak_password"))
		return
	}
	_, _ = h.S.Register(in.Name, in.Email, in.Password)
	WriteJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var in reqRefresh
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	claims, err := security.Parse(h.S.JWTSecret, in.Refresh)
	if err != nil {
		WriteAppError(w, errs.InvalidRefresh)
		return
	}
	sub, ok := claims["sub"].(float64)
	if !ok {
		WriteAppError(w, errs.InvalidRefresh)
		return
	}

	ua, ip := clientUA(r), clientIP(r)
	a, r2, err := h.S.Refresh(int64(sub), in.Refresh, ua, ip)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"token": a, "refresh": r2})
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	var in reqRefresh
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if err := h.S.Logout(UID(r), in.Refresh); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandlers) TotpSetup(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	email := r.URL.Query().Get("email")
	secret, url, err := h.S.TotpSetup(uid, email)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"secret": secret, "otpauth": url})
}

func (h *AuthHandlers) TotpConfirm(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var in reqTotpConfirm
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if err := h.S.TotpConfirm(uid, in.Code); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var in reqLogin
	if err := DecodeStrict(r, &in); err != nil {
		Fail(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}
	if err := validation.ValidateStruct(in); err != nil {
		WriteAppError(w, errs.ValidationFailed(validation.ValidationMessage(err)))
		return
	}
	ua, ip := clientUA(r), clientIP(r)
	access, refresh, uid, err := h.S.Login(in.Email, in.Password, in.DeviceID, in.DeviceName, in.Totp, ua, ip, in.Captcha)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"token": access, "refresh": refresh,
		"user": map[string]any{"id": uid, "email": in.Email},
	})
}

/* ---- Oturum görünürlüğü / yönetimi ---- */

func (h *AuthHandlers) Sessions(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	size, _ := strconv.Atoi(r.URL.Query().Get("size"))
	if size < 1 {
		size = 20
	}
	rows, total, err := h.S.Sessions(UID(r), page, size)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"total": total, "data": rows})
}

func (h *AuthHandlers) SessionDelete(w http.ResponseWriter, r *http.Request) {
	sid, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.S.RevokeSession(UID(r), sid); err != nil {
		FromError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
