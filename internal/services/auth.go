package services

import (
	"time"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/security"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type AuthService struct {
	Repo       ports.AuthRepo
	JWTSecret  []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
	Onboard    *OnboardService
	Audit      *AuditService
}

func (s *AuthService) Register(name, email, pass string) (int64, error) {
	hash, err := security.ArgonHash(pass)
	if err != nil {
		return 0, err
	}
	u := &ports.User{Name: name, Email: email, PassHash: hash}
	if err := s.Repo.CreateUser(u); err != nil {
		return 0, err
	}
	if s.Onboard != nil {
		_ = s.Onboard.Seed(u.ID)
	}
	if s.Audit != nil {
		s.Audit.Log(u.ID, "auth.register", "user", &u.ID, map[string]any{"email": email})
	}
	return u.ID, nil
}

func (s *AuthService) Login(email, pass, deviceID, deviceName, totpCode string) (access, refresh string, userID int64, err error) {
	u, err := s.Repo.FindUserByEmail(email)
	if err != nil {
		return "", "", 0, errs.InvalidCredentials
	}
	ok, _ := security.ArgonCheck(pass, u.PassHash)
	if !ok {
		return "", "", 0, errs.InvalidCredentials
	}

	if ts, _ := s.Repo.GetTotp(u.ID); ts != nil && ts.ConfirmedAt != nil {
		if totpCode == "" {
			return "", "", 0, errs.TOTPRequired
		}
		if !totp.Validate(totpCode, ts.Secret) {
			return "", "", 0, errs.TOTPInvalid
		}
	}

	access, err = security.SignAccess(s.JWTSecret, u.ID, s.AccessTTL)
	if err != nil {
		return
	}
	refresh, err = security.SignRefresh(s.JWTSecret, u.ID, s.RefreshTTL)
	if err != nil {
		return
	}
	_ = s.Repo.StoreRefresh(u.ID, security.SHA256Hex(refresh), time.Now().Add(s.RefreshTTL))

	if deviceID != "" {
		_ = s.Repo.UpsertDevice(u.ID, deviceID, deviceName, time.Now())
	}
	if s.Audit != nil {
		s.Audit.Log(u.ID, "auth.login", "user", &u.ID, map[string]any{"deviceId": deviceID, "deviceName": deviceName})
	}
	return access, refresh, u.ID, nil
}

func (s *AuthService) Refresh(userID int64, oldRefresh string) (newAccess, newRefresh string, err error) {
	claims, err := security.Parse(s.JWTSecret, oldRefresh)
	if err != nil {
		return "", "", errs.InvalidRefresh
	}
	if typ, _ := claims["typ"].(string); typ != "refresh" {
		return "", "", errs.InvalidRefresh
	}
	sub, _ := claims["sub"].(float64)
	if int64(sub) != userID {
		return "", "", errs.InvalidRefresh
	}

	ok, err := s.Repo.HasValidRefresh(userID, security.SHA256Hex(oldRefresh), time.Now())
	if err != nil || !ok {
		return "", "", errs.InvalidRefresh
	}

	newAccess, err = security.SignAccess(s.JWTSecret, userID, s.AccessTTL)
	if err != nil {
		return
	}
	newRefresh, err = security.SignRefresh(s.JWTSecret, userID, s.RefreshTTL)
	if err != nil {
		return
	}
	err = s.Repo.RotateRefresh(userID, security.SHA256Hex(oldRefresh), security.SHA256Hex(newRefresh), time.Now().Add(s.RefreshTTL))
	return
}

func (s *AuthService) Logout(userID int64, refresh string) error {
	if s.Audit != nil {
		s.Audit.Log(userID, "auth.logout", "user", &userID, nil)
	}
	return s.Repo.InvalidateRefresh(userID, security.SHA256Hex(refresh))
}

func (s *AuthService) TotpSetup(userID int64, email string) (secret, otpauth string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer: s.Issuer, AccountName: email, Period: 30, Digits: otp.DigitsSix,
	})
	if err != nil {
		return "", "", err
	}
	secret = key.Secret()
	if err = s.Repo.SetTotp(userID, secret); err != nil {
		return "", "", err
	}
	if s.Audit != nil {
		s.Audit.Log(userID, "auth.totp.setup", "user", &userID, nil)
	}
	return secret, key.URL(), nil
}

func (s *AuthService) TotpConfirm(userID int64, code string) error {
	ts, err := s.Repo.GetTotp(userID)
	if err != nil || ts == nil {
		return errs.ValidationFailed("no totp")
	}
	if !totp.Validate(code, ts.Secret) {
		return errs.TOTPInvalid
	}
	if s.Audit != nil {
		s.Audit.Log(userID, "auth.totp.confirm", "user", &userID, nil)
	}
	return s.Repo.ConfirmTotp(userID)
}
