package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	captchaadp "github.com/Veysel440/finance-master-api/internal/adapters/captcha"
	mysqladp "github.com/Veysel440/finance-master-api/internal/adapters/mysql"
	ratesadp "github.com/Veysel440/finance-master-api/internal/adapters/rates"
	"github.com/Veysel440/finance-master-api/internal/config"
	"github.com/Veysel440/finance-master-api/internal/cron"
	apihttp "github.com/Veysel440/finance-master-api/internal/http"
	"github.com/Veysel440/finance-master-api/internal/obs"
	"github.com/Veysel440/finance-master-api/internal/services"
)

func main() {
	time.Local = time.UTC

	obs.InitLogger()
	obs.InitMetrics()
	stopTrace := obs.InitTracing("finance-master-api")
	defer func() { _ = stopTrace(context.Background()) }()

	cfg := config.Load()

	db, err := mysqladp.Open(cfg.DSN)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec("SET time_zone = '+00:00'"); err != nil {
		log.Println("warn: set time_zone failed:", err)
	}

	authRepo := mysqladp.NewAuthRepo(db)
	txRepo := mysqladp.NewTxRepo(db)
	walletRepo := mysqladp.NewWalletRepo(db)
	catRepo := mysqladp.NewCategoryRepo(db)
	auditRepo := mysqladp.NewAuditRepo(db)
	sessionRepo := mysqladp.NewSessionRepo(db)
	idemRepo := mysqladp.NewIdempotencyRepo(db.DB) // *sql.DB
	ratesStore := mysqladp.NewRatesStore(db.DB)    // *sql.DB

	auditSvc := &services.AuditService{Repo: auditRepo}
	onboard := &services.OnboardService{Wallet: walletRepo, Cat: catRepo}

	var capVerifier services.CaptchaVerifier
	switch strings.ToLower(cfg.CaptchaProvider) {
	case "turnstile":
		capVerifier = &captchaadp.Turnstile{Secret: cfg.TurnstileSecret}
	case "recaptcha":
		capVerifier = &captchaadp.ReCaptcha{Secret: cfg.RecaptchaSecret}
	}

	authSvc := &services.AuthService{
		Repo:       authRepo,
		Sess:       sessionRepo,
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

		CaptchaThreshold: cfg.CaptchaThreshold,
		Captcha:          capVerifier,
	}

	txSvc := &services.TxService{Repo: txRepo, Audit: auditSvc, Idem: idemRepo}
	walletSvc := &services.WalletService{Repo: walletRepo, Audit: auditSvc}
	catSvc := &services.CategoryService{Repo: catRepo, Audit: auditSvc}

	httpClient := &http.Client{Timeout: 8 * time.Second}
	ratesFetcher := &ratesadp.HTTPClient{BaseURL: cfg.RatesURL, Client: httpClient}
	ratesSvc := &services.RatesService{
		F:        ratesFetcher,
		Store:    ratesStore,
		TTL:      cfg.RatesTTL,
		StaleTTL: cfg.RatesStaleTTL,
	}

	if cfg.RatesWarmEvery > 0 {
		stop := cron.StartRatesWarm(context.Background(), ratesSvc, cfg.RatesWarmBases, cfg.RatesWarmEvery)
		defer stop()
	}

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
