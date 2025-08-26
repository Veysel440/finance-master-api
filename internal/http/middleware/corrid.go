package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// benzersiz context anahtarı
type ctxKeyReqID struct{}

var reqIDKey = ctxKeyReqID{}

// X-Request-Id üret ve yanıta yaz. Varsa geleni kullan.
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if id == "" {
			id = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), reqIDKey, id)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Yardımcı: request id al
func ReqID(r *http.Request) string {
	if v := r.Context().Value(reqIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
