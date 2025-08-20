package services

import (
	"errors"
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	Repo       ports.AuthRepo
	JWTSecret  []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func (s *AuthService) Register(name, email, pass string) error {
	hash, _ := argon2id.CreateHash(pass, argon2id.DefaultParams)
	u := &ports.User{Name: name, Email: email, PassHash: hash}
	return s.Repo.CreateUser(u)
}

func (s *AuthService) Login(email, pass string) (access, refresh string, err error) {
	u, err := s.Repo.FindUserByEmail(email)
	if err != nil {
		return "", "", err
	}
	ok, _ := argon2id.ComparePasswordAndHash(pass, u.PassHash)
	if !ok {
		return "", "", errors.New("invalid_credentials")
	}
	now := time.Now()

	accessTok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID, "exp": now.Add(s.AccessTTL).Unix(),
	})
	access, _ = accessTok.SignedString(s.JWTSecret)

	refreshTok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": u.ID, "typ": "refresh", "exp": now.Add(s.RefreshTTL).Unix(),
	})
	refresh, _ = refreshTok.SignedString(s.JWTSecret)
	_ = s.Repo.StoreRefresh(u.ID, refresh, now.Add(s.RefreshTTL))
	return
}
