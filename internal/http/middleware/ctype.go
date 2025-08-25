package middleware

import "net/http"

func RequireJSON() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost, http.MethodPut, http.MethodPatch:
				ct := r.Header.Get("Content-Type")
				if ct == "" || (ct != "application/json" && ct[:16] != "application/json") {
					w.WriteHeader(http.StatusUnsupportedMediaType)
					_, _ = w.Write([]byte(`{"error":"unsupported_media_type","code":415,"message":"Content-Type must be application/json"}`))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
