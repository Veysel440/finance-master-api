package http

import (
	"encoding/json"
	"net/http"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/security"
	"github.com/Veysel440/finance-master-api/internal/services"
)

type AuthHandlers struct{ S *services.AuthService }

type reqRegister struct{ Name, Email, Password string }
type reqLogin struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	Totp       string `json:"totp,omitempty"`
	DeviceID   string `json:"deviceId,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
}
type reqRefresh struct {
	Refresh string `json:"refresh"`
}
type reqTotpConfirm struct {
	Code string `json:"code"`
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var in reqRegister
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if in.Email == "" || in.Password == "" {
		WriteAppError(w, errs.ValidationFailed("email and password required"))
		return
	}
	uid, err := h.S.Register(in.Name, in.Email, in.Password)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 201, map[string]any{
		"status": "ok",
		"user":   map[string]any{"id": uid, "name": in.Name, "email": in.Email},
	})
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var in reqLogin
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	access, refresh, uid, err := h.S.Login(in.Email, in.Password, in.DeviceID, in.DeviceName, in.Totp)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, map[string]any{
		"token": access, "refresh": refresh,
		"user": map[string]any{"id": uid, "email": in.Email},
	})
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var in reqRefresh
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
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
	uid := int64(sub)

	a, r2, err := h.S.Refresh(uid, in.Refresh)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, map[string]string{"token": a, "refresh": r2})
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	var in reqRefresh
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if err := h.S.Logout(UID(r), in.Refresh); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *AuthHandlers) TotpSetup(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	email := r.URL.Query().Get("email")
	secret, url, err := h.S.TotpSetup(uid, email)
	if err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, map[string]string{"secret": secret, "otpauth": url})
}

func (h *AuthHandlers) TotpConfirm(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var in reqTotpConfirm
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		Fail(w, 400, "bad_request", "invalid json")
		return
	}
	if err := h.S.TotpConfirm(uid, in.Code); err != nil {
		FromError(w, err)
		return
	}
	WriteJSON(w, 200, map[string]string{"status": "ok"})
}
