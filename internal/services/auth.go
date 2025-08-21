package services

import (
	"errors"
	"time"

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
}

func (s *AuthService) Register(name, email, pass string) error {
	hash, err := security.ArgonHash(pass)
	if err != nil {
		return err
	}
	u := &ports.User{Name: name, Email: email, PassHash: hash}
	return s.Repo.CreateUser(u)
}

// Login with optional TOTP and device
func (s *AuthService) Login(email, pass, deviceID, deviceName, totpCode string) (access, refresh string, userID int64, err error) {
	u, err := s.Repo.FindUserByEmail(email)
	if err != nil {
		return "", "", 0, errors.New("invalid_credentials")
	}
	ok, _ := security.ArgonCheck(pass, u.PassHash)
	if !ok {
		return "", "", 0, errors.New("invalid_credentials")
	}

	if ts, _ := s.Repo.GetTotp(u.ID); ts != nil && ts.ConfirmedAt != nil {
		if totpCode == "" {
			return "", "", 0, errors.New("totp_required")
		}
		if !totp.Validate(totpCode, ts.Secret) {
			return "", "", 0, errors.New("totp_invalid")
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
	return access, refresh, u.ID, nil
}

func (s *AuthService) Refresh(userID int64, oldRefresh string) (newAccess, newRefresh string, err error) {

	claims, err := security.Parse(s.JWTSecret, oldRefresh)
	if err != nil {
		return "", "", errors.New("invalid_refresh")
	}
	if typ, _ := claims["typ"].(string); typ != "refresh" {
		return "", "", errors.New("invalid_refresh")
	}
	sub, _ := claims["sub"].(float64)
	if int64(sub) != userID {
		return "", "", errors.New("invalid_refresh")
	}

	ok, err := s.Repo.HasValidRefresh(userID, security.SHA256Hex(oldRefresh), time.Now())
	if err != nil || !ok {
		return "", "", errors.New("invalid_refresh")
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
	return secret, key.URL(), nil
}

func (s *AuthService) TotpConfirm(userID int64, code string) error {
	ts, err := s.Repo.GetTotp(userID)
	if err != nil || ts == nil {
		return errors.New("no_totp")
	}
	if !totp.Validate(code, ts.Secret) {
		return errors.New("totp_invalid")
	}
	return s.Repo.ConfirmTotp(userID)
}
