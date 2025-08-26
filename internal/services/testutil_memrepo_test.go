package services

import (
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/security"
)

type memAuthRepo struct {
	users   map[string]*ports.User
	refresh map[int64]map[string]time.Time
	totp    map[int64]*ports.TotpSecret
}

func newMemAuthRepo() *memAuthRepo {
	return &memAuthRepo{
		users:   map[string]*ports.User{},
		refresh: map[int64]map[string]time.Time{},
		totp:    map[int64]*ports.TotpSecret{},
	}
}

func mustHash(p string) string {
	h, _ := security.ArgonHash(p)
	return h
}

/* ports.AuthRepo */

func (m *memAuthRepo) CreateUser(u *ports.User) error {
	m.users[u.Email] = &ports.User{ID: u.ID, Email: u.Email, PassHash: u.PassHash, Name: u.Name}
	return nil
}
func (m *memAuthRepo) FindUserByEmail(email string) (*ports.User, error) {
	if u, ok := m.users[email]; ok {
		cp := *u
		return &cp, nil
	}
	return nil, ErrNotFound
}

var ErrNotFound = errString("not_found")

type errString string

func (e errString) Error() string { return string(e) }

func (m *memAuthRepo) StoreRefresh(userID int64, hash string, exp time.Time) error {
	if m.refresh[userID] == nil {
		m.refresh[userID] = map[string]time.Time{}
	}
	m.refresh[userID][hash] = exp
	return nil
}
func (m *memAuthRepo) InvalidateRefresh(userID int64, hash string) error {
	if m.refresh[userID] != nil {
		delete(m.refresh[userID], hash)
	}
	return nil
}
func (m *memAuthRepo) HasValidRefresh(userID int64, hash string, now time.Time) (bool, error) {
	if m.refresh[userID] == nil {
		return false, nil
	}
	exp, ok := m.refresh[userID][hash]
	if !ok {
		return false, nil
	}
	return now.Before(exp), nil
}
func (m *memAuthRepo) RotateRefresh(userID int64, oldHash, newHash string, newExp time.Time) error {
	_ = m.InvalidateRefresh(userID, oldHash)
	return m.StoreRefresh(userID, newHash, newExp)
}
func (m *memAuthRepo) UpsertDevice(_ int64, _ string, _ string, _ time.Time) error { return nil }

func (m *memAuthRepo) GetTotp(userID int64) (*ports.TotpSecret, error) {
	return m.totp[userID], nil
}
func (m *memAuthRepo) SetTotp(userID int64, secret string) error {
	m.totp[userID] = &ports.TotpSecret{UserID: userID, Secret: secret}
	return nil
}
func (m *memAuthRepo) ConfirmTotp(userID int64) error {
	ts := m.totp[userID]
	if ts == nil {
		return nil
	}
	now := time.Now()
	ts.ConfirmedAt = &now
	return nil
}
