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

/* Opsiyonel CAPTCHA doğrulayıcı */
type CaptchaVerifier interface {
	Verify(token, ip, ua string) bool
}

type AuthService struct {
	Repo       ports.AuthRepo
	Sess       ports.SessionRepo
	JWTSecret  []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string
	Onboard    *OnboardService
	Audit      *AuditService

	// UA/IP binding ve brute-force koruması
	BindRefreshToUA bool
	BindRefreshToIP bool

	MaxLoginFailures int           // kilit eşiği
	LockFor          time.Duration // kilit süresi
	FailWindow       time.Duration // deneme penceresi

	BackoffBase      time.Duration // exponential backoff başlangıç
	BackoffCap       time.Duration // backoff üst sınır
	CaptchaThreshold int           // şu kadar ardışık fail → captcha
	Captcha          CaptchaVerifier

	// in-memory fallback
	mu      sync.Mutex
	failMem map[string][]time.Time // key=email|ip
	lockMem map[int64]time.Time    // userID → until
}

func (s *AuthService) defaults() {
	if s.MaxLoginFailures == 0 {
		s.MaxLoginFailures = 10
	}
	if s.LockFor == 0 {
		s.LockFor = 15 * time.Minute
	}
	if s.FailWindow == 0 {
		s.FailWindow = 10 * time.Minute
	}
	if s.BackoffBase == 0 {
		s.BackoffBase = 2 * time.Second
	}
	if s.BackoffCap == 0 {
		s.BackoffCap = 60 * time.Second
	}
	if s.CaptchaThreshold == 0 {
		s.CaptchaThreshold = 5
	}
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

func (s *AuthService) Login(email, pass, deviceID, deviceName, totpCode, ua, ip, captcha string) (access, refresh string, userID int64, err error) {
	s.defaults()

	u, err := s.Repo.FindUserByEmail(email)
	if err != nil {
		_, fails := s.bumpFail(0, email, ip)
		return "", "", 0, s.backoffOrInvalid(fails)
	}

	if until := s.getLock(u.ID); until != nil && time.Now().Before(*until) {
		return "", "", 0, errs.AccountLocked
	}

	if s.CaptchaThreshold > 0 && s.needCaptcha(email, ip) {
		if s.Captcha == nil || !s.Captcha.Verify(captcha, ip, ua) {
			return "", "", 0, errs.CaptchaRequired
		}
	}

	ok, _ := security.ArgonCheck(pass, u.PassHash)
	if !ok {
		locked, fails := s.bumpFail(u.ID, email, ip)
		if locked {
			return "", "", 0, errs.AccountLocked
		}
		return "", "", 0, s.backoffOrInvalid(fails)
	}

	if ts, _ := s.Repo.GetTotp(u.ID); ts != nil && ts.ConfirmedAt != nil {
		if totpCode == "" || !totp.Validate(totpCode, ts.Secret) {
			locked, fails := s.bumpFail(u.ID, email, ip)
			if locked {
				return "", "", 0, errs.AccountLocked
			}
			return "", "", 0, s.backoffOrInvalid(fails)
		}
	}

	// başarılı giriş
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
	exp := now.Add(s.RefreshTTL)

	if s.Sess != nil {
		_ = s.Sess.StoreRefreshMeta(u.ID, hash, ua, ip, exp, now)
	} else {
		_ = s.Repo.StoreRefresh(u.ID, hash, exp)
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
	if s.Sess != nil {
		valid, err = s.Sess.ValidateRefresh(userID, hash, ua, ip, now, s.BindRefreshToUA, s.BindRefreshToIP)
	} else {
		valid, err = s.Repo.HasValidRefresh(userID, hash, now)
	}
	if err != nil || !valid {
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

	newHash := security.SHA256Hex(newRefresh)
	exp := now.Add(s.RefreshTTL)

	if s.Sess != nil {
		err = s.Sess.RotateRefreshMeta(userID, hash, newHash, ua, ip, exp, now)
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

/* ---- Oturum listesi / iptali ---- */

func (s *AuthService) Sessions(userID int64) ([]ports.Session, error) {
	if s.Sess == nil {
		return nil, errs.Forbidden
	}
	return s.Sess.ListSessions(userID)
}

func (s *AuthService) RevokeSession(userID, sid int64) error {
	if s.Sess == nil {
		return errs.Forbidden
	}
	return s.Sess.RevokeSession(userID, sid)
}

/* ---- brute-force yardımcıları ---- */

func (s *AuthService) needCaptcha(email, ip string) bool {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failMem == nil {
		return false
	}
	cut := now.Add(-s.FailWindow)
	key := email + "|" + ip
	arr := s.failMem[key]
	n := 0
	for _, t := range arr {
		if t.After(cut) {
			n++
		}
	}
	return n >= s.CaptchaThreshold
}

func (s *AuthService) backoffOrInvalid(fails int) error {
	if fails <= 1 {
		return errs.InvalidCredentials
	}
	wait := s.BackoffBase
	steps := fails - 1
	for steps > 0 && wait < s.BackoffCap {
		wait *= 2
		steps--
		if wait > s.BackoffCap {
			wait = s.BackoffCap
			break
		}
	}
	return errs.SlowDownAfter(int(wait.Seconds()))
}

func (s *AuthService) bumpFail(userID int64, email, ip string) (locked bool, fails int) {
	now := time.Now()
	window := s.FailWindow
	max := s.MaxLoginFailures
	lock := s.LockFor

	// Kalıcı guard varsa onu kullan
	if lg, ok := s.Repo.(ports.LoginGuardRepo); ok {
		fc, until, _ := lg.IncLoginFail(email, ip, now, window)
		if until != nil && now.Before(*until) {
			return true, fc
		}
		if fc >= max {
			_ = lg.LockUser(userID, now.Add(lock))
			return true, fc
		}
		return false, fc
	}

	// In-memory fallback
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
	cut := now.Add(-window)
	trim := arr[:0]
	for _, t := range arr {
		if t.After(cut) {
			trim = append(trim, t)
		}
	}
	s.failMem[key] = trim
	fc := len(trim)

	if fc >= max {
		s.lockMem[userID] = now.Add(lock)
		return true, fc
	}
	if until, ok := s.lockMem[userID]; ok && now.Before(until) {
		return true, fc
	}
	return false, fc
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
		until, _ := lg.GetUserLock(userID)
		return until
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if until, ok := s.lockMem[userID]; ok {
		u := until
		return &u
	}
	return nil
}
