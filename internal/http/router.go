package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func Router(api *Handlers, secret []byte) http.Handler {
	r := chi.NewRouter()

	// Global Middlewares
	r.Use(cors.AllowAll().Handler)
	r.Use(httprate.LimitByIP(200, 1*60*1e9)) // 200 req/min

	r.Route("/v1", func(r chi.Router) {
		// Auth
		r.Post("/auth/register", api.Register)
		r.Post("/auth/login", api.Login)

		// Private Routes
		r.Group(func(pr chi.Router) {
			pr.Use(Auth(secret))

			// Transactions
			pr.Get("/transactions", api.TxList)
			pr.Post("/transactions", api.TxCreate)
			pr.Put("/transactions/{id}", api.TxUpdate)
			pr.Delete("/transactions/{id}", api.TxDelete)

			// Sync
			pr.Get("/sync/transactions", api.TxSince)
			pr.Post("/sync/transactions", api.TxUpsertBatch)
		})
	})

	return r
}
