package services

import (
	"testing"
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

func Test_Refresh_UA_IP_Binding(t *testing.T) {
	mem := newMemAuthRepo()
	_ = mem.CreateUser(&ports.User{ID: 1, Email: "u@x", PassHash: mustHash("Password123")})

	s := &AuthService{
		Repo:            mem,
		JWTSecret:       []byte("k"),
		AccessTTL:       time.Hour,
		RefreshTTL:      24 * time.Hour,
		BindRefreshToUA: true,
		BindRefreshToIP: true,
	}

	a, r, _, err := s.Login("u@x", "Password123", "", "", "", "uaA", "1.1.1.1", "")
	if err != nil || a == "" || r == "" {
		t.Fatalf("login failed: %v", err)
	}

	if _, _, err := s.Refresh(1, r, "uaB", "2.2.2.2"); err == nil {
		t.Fatal("expected invalid refresh")
	}
}
