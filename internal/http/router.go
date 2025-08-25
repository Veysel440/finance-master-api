package http

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Veysel440/finance-master-api/internal/errs"
	imw "github.com/Veysel440/finance-master-api/internal/http/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

/* -------- CORS: alan-adı beyaz listesi -------- */

func whitelistOrigins() []string {
	raw := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
	if raw == "" {

		return []string{
			"http://localhost:19006", "http://127.0.0.1:19006", // Expo dev
			"http://localhost:8081", "http://127.0.0.1:8081",
			"http://localhost:5173", "http://127.0.0.1:5173",
		}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func corsWhitelist() func(http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins:   whitelistOrigins(),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}).Handler
}

/* -------- Router -------- */

func Router(api *API) http.Handler {
	r := chi.NewRouter()

	// CORS (beyaz liste)
	r.Use(corsWhitelist())

	// Çekirdek middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(Common())               // Recoverer
	r.Use(imw.BodyLimit(1 << 20)) // 1 MiB gövde sınırı
	r.Use(imw.Idempotency())      // Idempotency-Key -> context

	// Global 404/405
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) { WriteAppError(w, errs.NotFound) })
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		Fail(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})

	// Sağlık
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// API
	r.Route("/v1", func(r chi.Router) {
		// ---- Auth uçları: daha sıkı limitler ----
		r.With(httprate.LimitByIP(20, time.Minute)).Post("/auth/register", api.Auth.Register)
		r.With(httprate.LimitByIP(10, time.Minute)).Post("/auth/login", api.Auth.Login)
		r.With(httprate.LimitByIP(30, time.Minute)).Post("/auth/refresh", api.Auth.Refresh)

		// ---- Korunan alan ----
		r.Group(func(pr chi.Router) {
			pr.Use(Auth(api.Secret))

			// Auth extras
			pr.With(httprate.LimitByIP(30, time.Minute)).Post("/auth/logout", api.Auth.Logout)
			pr.With(httprate.LimitByIP(10, time.Minute)).Post("/auth/totp/setup", api.Auth.TotpSetup)
			pr.With(httprate.LimitByIP(10, time.Minute)).Post("/auth/totp/confirm", api.Auth.TotpConfirm)

			// Oturum görünürlüğü
			pr.With(httprate.LimitByIP(60, time.Minute)).Get("/auth/sessions", api.Auth.Sessions)
			pr.With(httprate.LimitByIP(30, time.Minute)).Delete("/auth/sessions/{id}", api.Auth.SessionDelete)

			// Kurlar
			pr.With(httprate.LimitByIP(120, time.Minute)).Get("/rates/latest", api.Rates.Latest)

			// İşlemler
			pr.With(httprate.LimitByIP(240, time.Minute)).Get("/transactions", api.H.TxList)
			pr.With(httprate.LimitByIP(240, time.Minute)).Get("/transactions/{id}", api.H.TxGetOne)
			pr.With(httprate.LimitByIP(120, time.Minute)).Post("/transactions", api.H.TxCreate)
			pr.With(httprate.LimitByIP(120, time.Minute)).Put("/transactions/{id}", api.H.TxUpdate)
			pr.With(httprate.LimitByIP(120, time.Minute)).Delete("/transactions/{id}", api.H.TxDelete)
			pr.With(httprate.LimitByIP(120, time.Minute)).Get("/transactions/summary", api.H.TxSummary)

			// Senkronizasyon
			pr.With(httprate.LimitByIP(240, time.Minute)).Get("/sync/transactions", api.H.TxSince)
			pr.With(httprate.LimitByIP(120, time.Minute)).Post("/sync/transactions", api.H.TxUpsertBatch)

			// Cüzdanlar
			pr.With(httprate.LimitByIP(240, time.Minute)).Get("/wallets", api.CatH.WalletList)
			pr.With(httprate.LimitByIP(120, time.Minute)).Post("/wallets", api.CatH.WalletCreate)
			pr.With(httprate.LimitByIP(120, time.Minute)).Put("/wallets/{id}", api.CatH.WalletUpdate)
			pr.With(httprate.LimitByIP(120, time.Minute)).Delete("/wallets/{id}", api.CatH.WalletDelete)

			// Kategoriler
			pr.With(httprate.LimitByIP(240, time.Minute)).Get("/categories", api.CatH.CategoryList)
			pr.With(httprate.LimitByIP(120, time.Minute)).Post("/categories", api.CatH.CategoryCreate)
			pr.With(httprate.LimitByIP(120, time.Minute)).Put("/categories/{id}", api.CatH.CategoryUpdate)
			pr.With(httprate.LimitByIP(120, time.Minute)).Delete("/categories/{id}", api.CatH.CategoryDelete)
		})
	})

	return r
}
