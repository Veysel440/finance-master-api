package middleware

import (
	"context"
	"net/http"
)

type ctxKey string

const idemKey ctxKey = "idem"

func Idempotency() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			k := r.Header.Get("Idempotency-Key")
			if k == "" {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), idemKey, k)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func FromContext(r *http.Request) string {
	v := r.Context().Value(idemKey)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
