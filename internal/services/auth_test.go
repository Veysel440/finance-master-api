package services

import (
	"testing"
	"time"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/security"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type fakeAuthRepo struct {
	nextID        int64
	users         map[int64]*ports.User
	byEmail       map[string]*ports.User
	totp          map[int64]*ports.TotpSecret
	refreshByUser map[int64]map[string]time.Time // hash -> exp
	devices       map[int64]map[string]string
}

func newFakeRepo() *fakeAuthRepo {
	return &fakeAuthRepo{
		nextID:        1,
		users:         map[int64]*ports.User{},
		byEmail:       map[string]*ports.User{},
		totp:          map[int64]*ports.TotpSecret{},
		refreshByUser: map[int64]map[string]time.Time{},
		devices:       map[int64]map[string]string{},
	}
}

// ports.AuthRepo implementation
func (r *fakeAuthRepo) CreateUser(u *ports.User) error {
	u.ID = r.nextID
	r.nextID++
	cp := *u
	r.users[u.ID] = &cp
	r.byEmail[u.Email] = &cp
	return nil
}
func (r *fakeAuthRepo) FindUserByEmail(email string) (*ports.User, error) {
	if u, ok := r.byEmail[email]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, errs.NotFound
}
func (r *fakeAuthRepo) GetTotp(uid int64) (*ports.TotpSecret, error) {
	if ts, ok := r.totp[uid]; ok {
		cp := *ts
		return &cp, nil
	}
	return nil, nil
}
func (r *fakeAuthRepo) StoreRefresh(uid int64, hash string, exp time.Time) error {
	m, ok := r.refreshByUser[uid]
	if !ok {
		m = map[string]time.Time{}
		r.refreshByUser[uid] = m
	}
	m[hash] = exp
	return nil
}
func (r *fakeAuthRepo) HasValidRefresh(uid int64, hash string, now time.Time) (bool, error) {
	if m, ok := r.refreshByUser[uid]; ok {
		if exp, ok2 := m[hash]; ok2 && exp.After(now) {
			return true, nil
		}
	}
	return false, nil
}
func (r *fakeAuthRepo) RotateRefresh(uid int64, oldHash, newHash string, newExp time.Time) error {
	if m, ok := r.refreshByUser[uid]; ok {
		delete(m, oldHash)
		m[newHash] = newExp
		return nil
	}
	return errs.NotFound
}
func (r *fakeAuthRepo) InvalidateRefresh(uid int64, hash string) error {
	if m, ok := r.refreshByUser[uid]; ok {
		delete(m, hash)
	}
	return nil
}
func (r *fakeAuthRepo) UpsertDevice(uid int64, deviceID, deviceName string, _ time.Time) error {
	m, ok := r.devices[uid]
	if !ok {
		m = map[string]string{}
		r.devices[uid] = m
	}
	m[deviceID] = deviceName
	return nil
}
func (r *fakeAuthRepo) SetTotp(uid int64, secret string) error {
	r.totp[uid] = &ports.TotpSecret{UserID: uid, Secret: secret}
	return nil
}
func (r *fakeAuthRepo) ConfirmTotp(uid int64) error {
	if ts, ok := r.totp[uid]; ok {
		now := time.Now()
		ts.ConfirmedAt = &now
		return nil
	}
	return errs.NotFound
}

// compile-time check
var _ ports.AuthRepo = (*fakeAuthRepo)(nil)

func newSvc() (*AuthService, *fakeAuthRepo) {
	repo := newFakeRepo()
	return &AuthService{
		Repo:       repo,
		JWTSecret:  []byte("test-secret"),
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
		Issuer:     "finmaster",
	}, repo
}

func mustRegister(t *testing.T, s *AuthService, email, pass string) int64 {
	t.Helper()
	id, err := s.Register("Veysel", email, pass)
	if err != nil {
		t.Fatalf("register error: %v", err)
	}
	return id
}

func TestRegister_OK(t *testing.T) {
	s, r := newSvc()
	id := mustRegister(t, s, "v@e.com", "A1complex!")
	u := r.users[id]
	if u == nil || u.Email != "v@e.com" {
		t.Fatalf("user not stored")
	}
	if ok, _ := security.ArgonCheck("A1complex!", u.PassHash); !ok {
		t.Fatalf("password hash mismatch")
	}
}

func TestLogin_OK_NoTOTP(t *testing.T) {
	s, r := newSvc()
	id := mustRegister(t, s, "v@e.com", "A1complex!")
	access, refresh, uid, err := s.Login("v@e.com", "A1complex!", "dev-1", "Pixel", "")
	if err != nil || access == "" || refresh == "" || uid != id {
		t.Fatalf("login failed: %v", err)
	}
	if len(r.refreshByUser[id]) == 0 {
		t.Fatalf("refresh not stored")
	}
}

func TestLogin_TOTPRequired_And_Invalid(t *testing.T) {
	s, _ := newSvc()
	id := mustRegister(t, s, "v@e.com", "A1complex!")
	// enable totp
	key, _ := totp.Generate(totp.GenerateOpts{Issuer: s.Issuer, AccountName: "v@e.com", Period: 30, Digits: otp.DigitsSix})
	_ = s.Repo.SetTotp(id, key.Secret())
	_ = s.Repo.ConfirmTotp(id)

	_, _, _, err := s.Login("v@e.com", "A1complex!", "dev", "X", "")
	if err != errs.TOTPRequired {
		t.Fatalf("expected TOTPRequired, got %v", err)
	}

	_, _, _, err = s.Login("v@e.com", "A1complex!", "dev", "X", "000000")
	if err != errs.TOTPInvalid {
		t.Fatalf("expected TOTPInvalid, got %v", err)
	}
}

func TestLogin_TOTPOK(t *testing.T) {
	s, _ := newSvc()
	id := mustRegister(t, s, "v@e.com", "A1complex!")
	key, _ := totp.Generate(totp.GenerateOpts{Issuer: s.Issuer, AccountName: "v@e.com", Period: 30, Digits: otp.DigitsSix})
	_ = s.Repo.SetTotp(id, key.Secret())
	_ = s.Repo.ConfirmTotp(id)
	code, _ := totp.GenerateCode(key.Secret(), time.Now())

	a, r, uid, err := s.Login("v@e.com", "A1complex!", "dev", "X", code)
	if err != nil || a == "" || r == "" || uid != id {
		t.Fatalf("login with totp failed: %v", err)
	}
}

func TestRefresh_OK_Rotates(t *testing.T) {
	s, repo := newSvc()
	id := mustRegister(t, s, "v@e.com", "A1complex!")
	_, refresh, _, _ := s.Login("v@e.com", "A1complex!", "", "", "")

	newA, newR, err := s.Refresh(id, refresh)
	if err != nil || newA == "" || newR == "" {
		t.Fatalf("refresh failed: %v", err)
	}

	oldHash := security.SHA256Hex(refresh)
	ok, _ := repo.HasValidRefresh(id, oldHash, time.Now())
	if ok {
		t.Fatalf("old refresh should be rotated out")
	}
}

func TestLogout_InvalidateRefresh(t *testing.T) {
	s, repo := newSvc()
	id := mustRegister(t, s, "v@e.com", "A1complex!")
	_, refresh, _, _ := s.Login("v@e.com", "A1complex!", "", "", "")
	hash := security.SHA256Hex(refresh)

	if err := s.Logout(id, refresh); err != nil {
		t.Fatalf("logout err: %v", err)
	}
	ok, _ := repo.HasValidRefresh(id, hash, time.Now())
	if ok {
		t.Fatalf("refresh should be invalidated")
	}
}

func TestTotpSetupConfirm_OK(t *testing.T) {
	s, _ := newSvc()
	id := mustRegister(t, s, "v@e.com", "A1complex!")
	secret, _, err := s.TotpSetup(id, "v@e.com")
	if err != nil || secret == "" {
		t.Fatalf("setup err: %v", err)
	}
	code, _ := totp.GenerateCode(secret, time.Now())
	if err := s.TotpConfirm(id, code); err != nil {
		t.Fatalf("confirm err: %v", err)
	}
}
