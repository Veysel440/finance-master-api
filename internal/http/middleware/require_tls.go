package middleware

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/Veysel440/finance-master-api/internal/errs"
)

// local writer to avoid import cycle with http package
func writeJSONErr(w http.ResponseWriter, e *errs.AppError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.HTTP)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error":   true,
		"code":    e.Code,
		"message": e.Message,
	})
}

// RequireTLS: TLS zorunlu. Proxy arkasında X-Forwarded-Proto / X-Forwarded-Ssl kontrol eder.
// Dev için ALLOW_HTTP=true ile esnetilebilir. /health muaf.
func RequireTLS(next http.Handler) http.Handler {
	allowHTTP := strings.EqualFold(os.Getenv("ALLOW_HTTP"), "true")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" || allowHTTP {
			next.ServeHTTP(w, r)
			return
		}
		if r.TLS != nil {
			next.ServeHTTP(w, r)
			return
		}
		if p := r.Header.Get("X-Forwarded-Proto"); strings.EqualFold(p, "https") {
			next.ServeHTTP(w, r)
			return
		}
		if s := r.Header.Get("X-Forwarded-Ssl"); strings.EqualFold(s, "on") {
			next.ServeHTTP(w, r)
			return
		}
		writeJSONErr(w, errs.InsecureTransport)
	})
}
