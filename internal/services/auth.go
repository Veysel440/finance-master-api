package services

import (
	"sync"
	"time"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/security"
	"github.com/Veysel440/finance-master-api/internal/validation"
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

	// UA/IP binding ve brute-force guard
	BindRefreshToUA bool
	BindRefreshToIP bool

	MaxLoginFailures int           // örn 10
	LockFor          time.Duration // örn 15m
	FailWindow       time.Duration // örn 10m

	// Repo yoksa in-memory yedek (tek süreç için)
	mu      sync.Mutex
	failMem map[string][]time.Time // key=email|ip
	lockMem map[int64]time.Time    // userID -> until
}

func (s *AuthService) Register(name, email, pass string) (int64, error) {
	if err := validation.ValidatePassword(pass, email); err != nil {
		return 0, errs.ValidationFailed("weak_password")
	}
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

func (s *AuthService) Login(email, pass, deviceID, deviceName, totpCode, ua, ip string) (access, refresh string, userID int64, err error) {
	u, err := s.Repo.FindUserByEmail(email)
	if err != nil {
		_ = s.bumpFail(0, email, ip)
		return "", "", 0, errs.InvalidCredentials
	}
	// Hesap kilidi
	if until := s.getLock(u.ID); until != nil && time.Now().Before(*until) {
		return "", "", 0, errs.AccountLocked
	}

	ok, _ := security.ArgonCheck(pass, u.PassHash)
	if !ok {
		if s.bumpFail(u.ID, email, ip) {
			return "", "", 0, errs.AccountLocked
		}
		return "", "", 0, errs.InvalidCredentials
	}

	// TOTP
	if ts, _ := s.Repo.GetTotp(u.ID); ts != nil && ts.ConfirmedAt != nil {
		if totpCode == "" {
			_ = s.bumpFail(u.ID, email, ip)
			return "", "", 0, errs.TOTPRequired
		}
		if !totp.Validate(totpCode, ts.Secret) {
			if s.bumpFail(u.ID, email, ip) {
				return "", "", 0, errs.AccountLocked
			}
			return "", "", 0, errs.TOTPInvalid
		}
	}

	// Başarılı giriş → sayaç reset
	s.resetFail(email, ip)
	if s.Audit != nil {
		s.Audit.Log(u.ID, "auth.login", "user", &u.ID, map[string]any{"deviceId": deviceID, "deviceName": deviceName, "ip": ip})
	}

	access, err = security.SignAccess(s.JWTSecret, u.ID, s.AccessTTL)
	if err != nil {
		return
	}
	refresh, err = security.SignRefresh(s.JWTSecret, u.ID, s.RefreshTTL)
	if err != nil {
		return
	}

	now := time.Now()
	hash := security.SHA256Hex(refresh)
	if sr, ok := s.Repo.(ports.SessionRepo); ok {
		_ = sr.StoreRefreshMeta(u.ID, hash, ua, ip, now.Add(s.RefreshTTL), now)
	} else {
		_ = s.Repo.StoreRefresh(u.ID, hash, now.Add(s.RefreshTTL))
	}
	if deviceID != "" {
		_ = s.Repo.UpsertDevice(u.ID, deviceID, deviceName, now)
	}
	return access, refresh, u.ID, nil
}

func (s *AuthService) Refresh(userID int64, oldRefresh, ua, ip string) (newAccess, newRefresh string, err error) {
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

	now := time.Now()
	hash := security.SHA256Hex(oldRefresh)
	valid := false
	if sr, ok := s.Repo.(ports.SessionRepo); ok {
		valid, err = sr.ValidateRefresh(userID, hash, ua, ip, now, s.BindRefreshToUA, s.BindRefreshToIP)
		if err != nil {
			return "", "", errs.InvalidRefresh
		}
		if !valid {
			return "", "", errs.SessionMismatch
		}
	} else {
		valid, err = s.Repo.HasValidRefresh(userID, hash, now)
		if err != nil || !valid {
			return "", "", errs.InvalidRefresh
		}
	}

	newAccess, err = security.SignAccess(s.JWTSecret, userID, s.AccessTTL)
	if err != nil {
		return
	}
	newRefresh, err = security.SignRefresh(s.JWTSecret, userID, s.RefreshTTL)
	if err != nil {
		return
	}

	newHash := security.SHA256Hex(newRefresh)
	exp := now.Add(s.RefreshTTL)
	if sr, ok := s.Repo.(ports.SessionRepo); ok {
		err = sr.RotateRefreshMeta(userID, hash, newHash, ua, ip, exp, now)
	} else {
		err = s.Repo.RotateRefresh(userID, hash, newHash, exp)
	}
	if err != nil {
		return "", "", errs.InvalidRefresh
	}
	return
}

func (s *AuthService) Logout(userID int64, refresh string) error {
	if s.Audit != nil {
		s.Audit.Log(userID, "auth.logout", "user", &userID, nil)
	}
	return s.Repo.InvalidateRefresh(userID, security.SHA256Hex(refresh))
}

func (s *AuthService) TotpSetup(userID int64, email string) (secret, otpauth string, err error) {
	key, err := totp.Generate(totp.GenerateOpts{Issuer: s.Issuer, AccountName: email, Period: 30, Digits: otp.DigitsSix})
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

// Oturum görünürlüğü
func (s *AuthService) Sessions(userID int64) ([]ports.Session, error) {
	if sr, ok := s.Repo.(ports.SessionRepo); ok {
		return sr.ListSessions(userID)
	}
	return []ports.Session{}, nil
}
func (s *AuthService) RevokeSession(userID, sessionID int64) error {
	if sr, ok := s.Repo.(ports.SessionRepo); ok {
		return sr.InvalidateSession(userID, sessionID)
	}
	return errs.NotFound
}

/* ---------- brute-force helper ---------- */

func (s *AuthService) bumpFail(userID int64, email, ip string) (locked bool) {
	now := time.Now()
	window := s.FailWindow
	if window == 0 {
		window = 10 * time.Minute
	}
	max := s.MaxLoginFailures
	if max == 0 {
		max = 10
	}
	lock := s.LockFor
	if lock == 0 {
		lock = 15 * time.Minute
	}

	if lg, ok := s.Repo.(ports.LoginGuardRepo); ok {
		fc, until, _ := lg.IncLoginFail(email, ip, now, window)
		if until != nil && now.Before(*until) {
			return true
		}
		if fc >= max {
			_ = lg.LockUser(userID, now.Add(lock))
			return true
		}
		return false
	}

	// in-memory fallback
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failMem == nil {
		s.failMem = map[string][]time.Time{}
	}
	if s.lockMem == nil {
		s.lockMem = map[int64]time.Time{}
	}

	key := email + "|" + ip
	arr := append(s.failMem[key], now)
	// window dışındakileri at
	trim := arr[:0]
	for _, t := range arr {
		if now.Sub(t) <= window {
			trim = append(trim, t)
		}
	}
	s.failMem[key] = trim
	if len(trim) >= max {
		s.lockMem[userID] = now.Add(lock)
		return true
	}
	if until, ok := s.lockMem[userID]; ok && now.Before(until) {
		return true
	}
	return false
}

func (s *AuthService) resetFail(email, ip string) {
	if lg, ok := s.Repo.(ports.LoginGuardRepo); ok {
		_ = lg.ResetLoginFail(email, ip)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failMem != nil {
		delete(s.failMem, email+"|"+ip)
	}
}

func (s *AuthService) getLock(userID int64) *time.Time {
	if lg, ok := s.Repo.(ports.LoginGuardRepo); ok {
		if until, _ := lg.GetUserLock(userID); until != nil {
			return until
		}
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if until, ok := s.lockMem[userID]; ok {
		u := until
		return &u
	}
	return nil
}
