package services

import (
	"testing"
	"time"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/security"
)

type fakeRepo struct{ ports.AuthRepo }

func Test_Login_BackoffAndLock(t *testing.T) {
	mem := newMemAuthRepo()
	pw, _ := security.ArgonHash("Password123")
	_ = mem.CreateUser(&ports.User{ID: 1, Email: "u@x", PassHash: pw})

	s := &AuthService{
		Repo:             mem,
		JWTSecret:        []byte("k"),
		AccessTTL:        time.Hour,
		RefreshTTL:       24 * time.Hour,
		MaxLoginFailures: 2,
		LockFor:          time.Minute,
		FailWindow:       time.Minute * 5,
		BackoffBase:      time.Second,
		BackoffCap:       time.Second * 2,
	}

	if _, _, _, err := s.Login("u@x", "bad", "", "", "", "ua", "ip", ""); err == nil {
		t.Fatal("want error")
	}

	if _, _, _, err := s.Login("u@x", "bad", "", "", "", "ua", "ip", ""); err != errs.AccountLocked {
		t.Fatalf("want locked, got %v", err)
	}
}
