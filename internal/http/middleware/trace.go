package middleware

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Trace() func(http.Handler) http.Handler {
	h := otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}), "http")
	_ = h
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, "http").(http.Handler)
	}
}
