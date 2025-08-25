package main

import (
	"context"
	"log"
	"net/http"
	"time"

	mysqladp "github.com/Veysel440/finance-master-api/internal/adapters/mysql"
	ratesadp "github.com/Veysel440/finance-master-api/internal/adapters/rates"
	"github.com/Veysel440/finance-master-api/internal/config"
	apihttp "github.com/Veysel440/finance-master-api/internal/http"
	"github.com/Veysel440/finance-master-api/internal/obs"
	"github.com/Veysel440/finance-master-api/internal/services"
)

func main() {

	time.Local = time.UTC

	// Observability
	obs.InitLogger()
	obs.InitMetrics()
	stopTrace := obs.InitTracing("finance-master-api")
	defer func() { _ = stopTrace(context.Background()) }()

	cfg := config.Load()

	// DB (sqlx.DB)
	db, err := mysqladp.Open(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	// Session TZ → UTC
	if _, err := db.Exec("SET time_zone = '+00:00'"); err != nil {
		log.Println("warn: set time_zone failed:", err)
	}

	authRepo := mysqladp.NewAuthRepo(db)
	txRepo := mysqladp.NewTxRepo(db)
	walletRepo := mysqladp.NewWalletRepo(db)
	catRepo := mysqladp.NewCategoryRepo(db)
	auditRepo := mysqladp.NewAuditRepo(db)

	idemRepo := mysqladp.NewIdempotencyRepo(db.DB)
	ratesStore := mysqladp.NewRatesStore(db.DB)

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

		BindRefreshToUA:  true,
		BindRefreshToIP:  true,
		MaxLoginFailures: 10,
		LockFor:          15 * time.Minute,
		FailWindow:       10 * time.Minute,
	}

	txSvc := &services.TxService{Repo: txRepo, Audit: auditSvc, Idem: idemRepo}
	walletSvc := &services.WalletService{Repo: walletRepo, Audit: auditSvc}
	catSvc := &services.CategoryService{Repo: catRepo, Audit: auditSvc}

	// Rates fetcher + persistent store fallback
	httpClient := &http.Client{Timeout: 8 * time.Second}
	ratesFetcher := &ratesadp.HTTPClient{BaseURL: cfg.RatesURL, Client: httpClient}
	ratesSvc := &services.RatesService{
		F:        ratesFetcher,
		Store:    ratesStore,
		TTL:      cfg.RatesTTL,
		StaleTTL: cfg.RatesStaleTTL,
	}

	// Handlers
	api := &apihttp.API{
		Auth:   &apihttp.AuthHandlers{S: authSvc},
		H:      &apihttp.Handlers{Auth: authSvc, Tx: txSvc},
		CatH:   &apihttp.CatalogHandlers{Wallet: walletSvc, Cat: catSvc},
		Rates:  &apihttp.RatesHandlers{S: ratesSvc},
		Secret: []byte(cfg.JWTSecret),
	}

	r := apihttp.Router(api)

	addr := cfg.Addr
	if addr == "" {
		addr = ":8080"
	}
	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
