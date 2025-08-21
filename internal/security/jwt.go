package security

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func SignAccess(secret []byte, userID int64, ttl time.Duration) (string, error) {
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "typ": "access", "exp": now.Add(ttl).Unix(),
	})
	return t.SignedString(secret)
}

func SignRefresh(secret []byte, userID int64, ttl time.Duration) (string, error) {
	now := time.Now()
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID, "typ": "refresh", "exp": now.Add(ttl).Unix(),
	})
	return t.SignedString(secret)
}

func Parse(secret []byte, token string) (jwt.MapClaims, error) {
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) { return secret, nil })
	return claims, err
}
