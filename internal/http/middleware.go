package http

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const userKey ctxKey = "uid"

func Auth(secret []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if len(h) < 8 {
				http.Error(w, "unauthorized", 401)
				return
			}
			tok := h[7:]
			claims := jwt.MapClaims{}
			_, err := jwt.ParseWithClaims(tok, claims, func(t *jwt.Token) (interface{}, error) { return secret, nil })
			if err != nil {
				http.Error(w, "unauthorized", 401)
				return
			}
			uid, ok := claims["sub"].(float64)
			if !ok {
				http.Error(w, "unauthorized", 401)
				return
			}
			ctx := context.WithValue(r.Context(), userKey, int64(uid))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
func UID(r *http.Request) int64 {
	v := r.Context().Value(userKey)
	if v == nil {
		return 0
	}
	return v.(int64)
}
