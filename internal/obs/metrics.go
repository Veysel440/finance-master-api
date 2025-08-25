package obs

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	Reg             = prometheus.NewRegistry()
	httpReqsTotal   = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total"}, []string{"method", "route", "status"})
	httpReqDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "http_request_duration_seconds", Buckets: prometheus.DefBuckets}, []string{"method", "route"})
)

func InitMetrics() {
	Reg.MustRegister(httpReqsTotal, httpReqDuration)
}

func MetricsHandler() http.Handler { return promhttp.HandlerFor(Reg, promhttp.HandlerOpts{}) }

func HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr := &statusRecorder{ResponseWriter: w, code: 200}
		start := time.Now()
		next.ServeHTTP(rr, r)
		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = r.URL.Path
		}
		httpReqsTotal.WithLabelValues(r.Method, route, http.StatusText(rr.code)).Inc()
		httpReqDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
	})
}

type statusRecorder struct {
	http.ResponseWriter
	code int
}

func (s *statusRecorder) WriteHeader(code int) { s.code = code; s.ResponseWriter.WriteHeader(code) }
