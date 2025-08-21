package main

import (
	"log"
	"net/http"

	"github.com/Veysel440/finance-master-api/internal/adapters/mysql"
	"github.com/Veysel440/finance-master-api/internal/config"
	h "github.com/Veysel440/finance-master-api/internal/http"
	"github.com/Veysel440/finance-master-api/internal/services"
)

func main() {
	cfg := config.Load()
	db, err := mysql.Open(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}

	authRepo := mysql.NewAuthRepo(db)
	txRepo := mysql.NewTxRepo(db)

	authSvc := &services.AuthService{
		Repo: authRepo, JWTSecret: []byte(cfg.JWTSecret),
		AccessTTL: cfg.AccessTTL, RefreshTTL: cfg.RefreshTTL, Issuer: "FinanceMaster",
	}
	txSvc := &services.TxService{Repo: txRepo}

	api := &h.API{
		Auth:   &h.AuthHandlers{S: authSvc},
		H:      &h.Handlers{Auth: authSvc, Tx: txSvc},
		Secret: []byte(cfg.JWTSecret),
	}

	r := h.Router(api)
	log.Println("listening on", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
