package main

import (
	"log"
	"net/http"

	"github.com/Veysel440/finance-master-api/internal/adapters/mysql"
	"github.com/Veysel440/finance-master-api/internal/config"
	h "github.com/Veysel440/finance-master-api/internal/http"
	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/Veysel440/finance-master-api/internal/services"
)

func main() {
	cfg := config.Load()
	_, err := mysql.Open(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: gerçek repo implementasyonlarını yaz
	var authRepo ports.AuthRepo
	var txRepo ports.TxRepo

	authSvc := &services.AuthService{Repo: authRepo, JWTSecret: []byte(cfg.JWTSecret), AccessTTL: cfg.AccessTTL, RefreshTTL: cfg.RefreshTTL}
	txSvc := &services.TxService{Repo: txRepo}
	handlers := &h.Handlers{Auth: authSvc, Tx: txSvc}

	r := h.Router(handlers, []byte(cfg.JWTSecret))
	log.Println("listening on", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
