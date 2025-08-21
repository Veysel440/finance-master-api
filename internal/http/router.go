package http

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"net/http"
)

type API struct {
	Auth   *AuthHandlers
	H      *Handlers
	Secret []byte
}

func Router(api *API) http.Handler {
	r := chi.NewRouter()
	r.Use(cors.AllowAll().Handler)
	r.Use(Common())
	r.Use(httprate.LimitByIP(200, 1*60*1e9))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte("ok")) })

	r.Route("/v1", func(r chi.Router) {
		r.Post("/auth/register", api.Auth.Register)
		r.Post("/auth/login", api.Auth.Login)

		r.Group(func(pr chi.Router) {
			pr.Use(Auth(api.Secret))

			pr.Post("/auth/refresh", api.Auth.Refresh)
			pr.Post("/auth/logout", api.Auth.Logout)

			pr.Post("/auth/totp/setup", api.Auth.TotpSetup)
			pr.Post("/auth/totp/confirm", api.Auth.TotpConfirm)

			pr.Get("/transactions", api.H.TxList)
			pr.Post("/transactions", api.H.TxCreate)
			pr.Put("/transactions/{id}", api.H.TxUpdate)
			pr.Delete("/transactions/{id}", api.H.TxDelete)
			pr.Get("/transactions/summary", api.H.TxSummary)

			pr.Get("/sync/transactions", api.H.TxSince)
			pr.Post("/sync/transactions", api.H.TxUpsertBatch)
		})
	})
	return r
}
