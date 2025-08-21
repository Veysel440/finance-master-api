package http

import (
	"encoding/json"
	"net/http"

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
		errJSON(w, 400, "bad_request")
		return
	}
	if in.Email == "" || in.Password == "" {
		errJSON(w, 400, "email_password_required")
		return
	}
	if err := h.S.Register(in.Name, in.Email, in.Password); err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 201, map[string]string{"status": "ok"})
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var in reqLogin
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		errJSON(w, 400, "bad_request")
		return
	}
	access, refresh, uid, err := h.S.Login(in.Email, in.Password, in.DeviceID, in.DeviceName, in.Totp)
	if err != nil {
		errJSON(w, 401, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"token": access, "refresh": refresh, "user": map[string]any{"id": uid, "email": in.Email}})
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var in reqRefresh
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		errJSON(w, 400, "bad_request")
		return
	}
	uid := UID(r)
	a, r2, err := h.S.Refresh(uid, in.Refresh)
	if err != nil {
		errJSON(w, 401, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"token": a, "refresh": r2})
}

func (h *AuthHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	var in reqRefresh
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		errJSON(w, 400, "bad_request")
		return
	}
	if err := h.S.Logout(UID(r), in.Refresh); err != nil {
		errJSON(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}

func (h *AuthHandlers) TotpSetup(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	email := r.URL.Query().Get("email")
	secret, url, err := h.S.TotpSetup(uid, email)
	if err != nil {
		errJSON(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"secret": secret, "otpauth": url})
}

func (h *AuthHandlers) TotpConfirm(w http.ResponseWriter, r *http.Request) {
	uid := UID(r)
	var in reqTotpConfirm
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		errJSON(w, 400, "bad_request")
		return
	}
	if err := h.S.TotpConfirm(uid, in.Code); err != nil {
		errJSON(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, map[string]string{"status": "ok"})
}
