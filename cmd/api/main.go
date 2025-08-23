package main

import (
	"log"
	"net/http"
	"time"

	mysqladp "github.com/Veysel440/finance-master-api/internal/adapters/mysql"
	ratesadp "github.com/Veysel440/finance-master-api/internal/adapters/rates"
	"github.com/Veysel440/finance-master-api/internal/config"
	h "github.com/Veysel440/finance-master-api/internal/http"
	"github.com/Veysel440/finance-master-api/internal/services"
)

func main() {
	cfg := config.Load()

	db, err := mysqladp.Open(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}

	// repos
	authRepo := mysqladp.NewAuthRepo(db)
	txRepo := mysqladp.NewTxRepo(db)
	walletRepo := mysqladp.NewWalletRepo(db)
	catRepo := mysqladp.NewCategoryRepo(db)
	auditRepo := mysqladp.NewAuditRepo(db)

	// services
	auditSvc := &services.AuditService{Repo: auditRepo}
	onboard := &services.OnboardService{Wallet: walletRepo, Cat: catRepo}

	authSvc := &services.AuthService{
		Repo:       authRepo,
		JWTSecret:  []byte(cfg.JWTSecret),
		AccessTTL:  cfg.AccessTTL,
		RefreshTTL: cfg.RefreshTTL,
		Issuer:     "FinanceMaster",
		Onboard:    onboard,
		Audit:      auditSvc,
	}
	txSvc := &services.TxService{Repo: txRepo, Audit: auditSvc}
	walletSvc := &services.WalletService{Repo: walletRepo, Audit: auditSvc}
	catSvc := &services.CategoryService{Repo: catRepo, Audit: auditSvc}

	ratesFetcher := &ratesadp.HTTPClient{BaseURL: cfg.RatesURL, Client: &http.Client{Timeout: 8 * time.Second}}
	ratesStore := &ratesadp.FileStore{Path: cfg.RatesCachePath}
	ratesSvc := &services.RatesService{F: ratesFetcher, Store: ratesStore, TTL: cfg.RatesTTL, StaleTTL: cfg.RatesStaleTTL}

	// handlers
	api := &h.API{
		Auth:   &h.AuthHandlers{S: authSvc},
		H:      &h.Handlers{Auth: authSvc, Tx: txSvc},
		CatH:   &h.CatalogHandlers{Wallet: walletSvc, Cat: catSvc},
		Rates:  &h.RatesHandlers{S: ratesSvc},
		Secret: []byte(cfg.JWTSecret),
	}

	r := h.Router(api)
	log.Println("listening on", cfg.Addr)
	log.Fatal(http.ListenAndServe(cfg.Addr, r))
}
