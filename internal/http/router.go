package http

import (
	"net/http"
	"time"

	"github.com/Veysel440/finance-master-api/internal/errs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func Router(api *API) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)
	r.Use(Common())
	r.Use(httprate.LimitByIP(200, time.Minute))

	// 404 / 405 JSON
	r.NotFound(func(w http.ResponseWriter, _ *http.Request) { WriteAppError(w, errs.NotFound) })
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		Fail(w, 405, "method_not_allowed", "method not allowed")
	})

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		WriteJSON(w, 200, map[string]string{"status": "ok"})
	})

	r.Route("/v1", func(r chi.Router) {
		r.Post("/auth/register", api.Auth.Register)
		r.Post("/auth/login", api.Auth.Login)
		r.Post("/auth/refresh", api.Auth.Refresh)

		r.Group(func(pr chi.Router) {
			pr.Use(Auth(api.Secret))

			pr.Post("/auth/logout", api.Auth.Logout)
			pr.Post("/auth/totp/setup", api.Auth.TotpSetup)
			pr.Post("/auth/totp/confirm", api.Auth.TotpConfirm)

			pr.Get("/rates/latest", api.Rates.Latest)

			pr.Get("/transactions", api.H.TxList)
			pr.Get("/transactions/{id}", api.H.TxGetOne)
			pr.Post("/transactions", api.H.TxCreate)
			pr.Put("/transactions/{id}", api.H.TxUpdate)
			pr.Delete("/transactions/{id}", api.H.TxDelete)
			pr.Get("/transactions/summary", api.H.TxSummary)

			pr.Get("/sync/transactions", api.H.TxSince)
			pr.Post("/sync/transactions", api.H.TxUpsertBatch)

			pr.Get("/wallets", api.CatH.WalletList)
			pr.Post("/wallets", api.CatH.WalletCreate)
			pr.Put("/wallets/{id}", api.CatH.WalletUpdate)
			pr.Delete("/wallets/{id}", api.CatH.WalletDelete)

			pr.Get("/categories", api.CatH.CategoryList)
			pr.Post("/categories", api.CatH.CategoryCreate)
			pr.Put("/categories/{id}", api.CatH.CategoryUpdate)
			pr.Delete("/categories/{id}", api.CatH.CategoryDelete)
		})
	})
	return r
}
