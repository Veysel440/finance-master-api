package middleware

import (
	"net/http"
	"time"

	"github.com/Veysel440/finance-master-api/internal/obs"
	"github.com/go-chi/chi/v5/middleware"
)

func RequestLog() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			start := time.Now()
			next.ServeHTTP(ww, r)
			d := time.Since(start)
			obs.Log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", ww.Status()).
				Str("req_id", middleware.GetReqID(r.Context())).
				Dur("dur", d).
				Msg("http")
		})
	}
}
