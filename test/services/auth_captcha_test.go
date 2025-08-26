package services_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/security"
	svc "github.com/Veysel440/finance-master-api/internal/services"
)

/* --- fakes --- */

type memAuthRepo struct {
	users map[string]*ports.User
}

func newMemAuthRepo() *memAuthRepo { return &memAuthRepo{users: map[string]*ports.User{}} }

func (m *memAuthRepo) CreateUser(u *ports.User) error {
	m.users[strings.ToLower(u.Email)] = u
	return nil
}
func (m *memAuthRepo) FindUserByEmail(email string) (*ports.User, error) {
	u := m.users[strings.ToLower(email)]
	if u == nil {
		return nil, fmt.Errorf("not found")
	}
	return u, nil
}
func (m *memAuthRepo) StoreRefresh(userID int64, refreshHash string, expires time.Time) error {
	return nil
}
func (m *memAuthRepo) InvalidateRefresh(userID int64, refreshHash string) error { return nil }
func (m *memAuthRepo) HasValidRefresh(userID int64, refreshHash string, now time.Time) (bool, error) {
	return true, nil
}
func (m *memAuthRepo) RotateRefresh(userID int64, oldHash, newHash string, newExp time.Time) error {
	return nil
}
func (m *memAuthRepo) UpsertDevice(userID int64, deviceID, name string, seen time.Time) error {
	return nil
}
func (m *memAuthRepo) GetTotp(userID int64) (*ports.TotpSecret, error) { return nil, nil }
func (m *memAuthRepo) SetTotp(userID int64, secret string) error       { return nil }
func (m *memAuthRepo) ConfirmTotp(userID int64) error                  { return nil }

func (m *memAuthRepo) IncLoginFail(email, ip string, now time.Time, window time.Duration) (int, *time.Time, error) {
	return 1, nil, nil
}
func (m *memAuthRepo) LockUser(userID int64, until time.Time) error { return nil }
func (m *memAuthRepo) ResetLoginFail(email, ip string) error        { return nil }
func (m *memAuthRepo) GetUserLock(userID int64) (*time.Time, error) { return nil, nil }

/* --- captcha fake --- */

type fakeCaptcha struct{ ok bool }

func (f fakeCaptcha) Verify(token, ip, ua string) bool { return f.ok }

func mustHash(p string) string {
	h, _ := security.ArgonHash(p)
	return h
}

func TestLogin_CaptchaRequiredAndPasses(t *testing.T) {
	mem := newMemAuthRepo()
	_ = mem.CreateUser(&ports.User{ID: 1, Email: "u@x.test", PassHash: mustHash("Password123")})

	s := &svc.AuthService{
		Repo:             mem,
		JWTSecret:        []byte("k"),
		AccessTTL:        time.Hour,
		RefreshTTL:       24 * time.Hour,
		CaptchaThreshold: 1,
	}

	_, _, _, _ = s.Login("u@x.test", "bad", "", "", "", "ua", "1.1.1.1", "")

	if _, _, _, err := s.Login("u@x.test", "Password123", "", "", "", "ua", "1.1.1.1", ""); err != errs.CaptchaRequired {
		t.Fatalf("expected captcha required, got %v", err)
	}

	s.Captcha = fakeCaptcha{ok: true}
	a, r, uid, err := s.Login("u@x.test", "Password123", "", "", "", "ua", "1.1.1.1", "tok")
	if err != nil || a == "" || r == "" || uid != 1 {
		t.Fatalf("login with captcha failed: %v", err)
	}
}
