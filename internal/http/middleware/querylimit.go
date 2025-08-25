package middleware

import "net/http"

func QueryLimit(maxLen int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, vals := range r.URL.Query() {
				for _, v := range vals {
					if len(v) > maxLen {
						w.WriteHeader(http.StatusBadRequest)
						_, _ = w.Write([]byte(`{"error":"validation_failed","code":400,"message":"query too long"}`))
						return
					}
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
