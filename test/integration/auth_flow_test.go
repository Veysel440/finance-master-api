//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/pressly/goose/v3"
	tc_mysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	apihttp "github.com/Veysel440/finance-master-api/internal/http"
	"github.com/Veysel440/finance-master-api/internal/services"
	// Adaptörlerin isimleri sende farklıysa aşağıyı güncelle:
	mysqladp "github.com/Veysel440/finance-master-api/internal/adapters/mysql"
)

func startMySQL(t *testing.T) (dsn string, terminate func()) {
	t.Helper()
	ctx := context.Background()
	c, err := tc_mysql.RunContainer(
		ctx,
		tc_mysql.WithDatabase("finance_master"),
		tc_mysql.WithUsername("root"),
		tc_mysql.WithPassword("testpass"),
	)
	if err != nil {
		t.Fatalf("container start: %v", err)
	}
	host, _ := c.Host(ctx)
	port, _ := c.MappedPort(ctx, "3306")
	dsn = fmt.Sprintf("root:testpass@tcp(%s:%s)/finance_master?parseTime=true&multiStatements=true", host, port.Port())
	return dsn, func() { _ = c.Terminate(ctx) }
}

func openDB(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// hazır olana kadar bekle
	deadline := time.Now().Add(30 * time.Second)
	for {
		if err = db.Ping(); err == nil || time.Now().After(deadline) {
			break
		}
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		t.Fatalf("ping db: %v", err)
	}
	return db
}

func migrate(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := goose.SetDialect("mysql"); err != nil {
		t.Fatalf("goose dialect: %v", err)
	}
	// Çalışma dizini modül kökü ise "migrations" doğru olur.
	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

type dummyRates struct{}

func (dummyRates) Latest(base string) (string, time.Time, map[string]float64, error) {
	return "TRY", time.Now(), map[string]float64{"USD": 1.0, "EUR": 0.9}, nil
}

func TestAuth_Flow_Register_Login_Refresh_Protected(t *testing.T) {
	dsn, stop := startMySQL(t)
	defer stop()

	db := openDB(t, dsn)
	defer db.Close()

	migrate(t, db)

	// Gerçek repos
	authRepo := mysqladp.NewAuthRepo(db)
	walRepo := mysqladp.NewWalletRepo(db)
	catRepo := mysqladp.NewCategoryRepo(db)

	// Servisler
	secret := []byte("test-secret")
	authSvc := &services.AuthService{
		Repo:       authRepo,
		JWTSecret:  secret,
		AccessTTL:  time.Hour,
		RefreshTTL: 24 * time.Hour,
		Issuer:     "finmaster",
		Onboard:    &services.OnboardService{Wallet: walRepo, Cat: catRepo},
		Audit:      &services.AuditService{Repo: mysqladp.NewAuditRepo(db)},
	}
	ratesSvc := &services.RatesService{F: dummyRates{}, TTL: time.Minute, StaleTTL: time.Hour}

	// Handler set
	api := &apihttp.API{
		Auth:   &apihttp.AuthHandlers{S: authSvc},
		H:      &apihttp.Handlers{}, // bu testte tx kullanılmıyor
		CatH:   &apihttp.CatalogHandlers{Wallet: &services.WalletService{Repo: walRepo}, Cat: &services.CategoryService{Repo: catRepo}},
		Rates:  &apihttp.RatesHandlers{S: ratesSvc},
		Secret: secret,
	}
	r := apihttp.Router(api)

	// 1) Register
	regBody := []byte(`{"name":"Veysel","email":"v@e.com","password":"A1complex!"}`)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewReader(regBody))
	r.ServeHTTP(w, req)
	if w.Code != 201 && w.Code != 200 {
		t.Fatalf("register status=%d body=%s", w.Code, w.Body.String())
	}

	// 2) Login
	loginBody := []byte(`{"email":"v@e.com","password":"A1complex!","deviceId":"dev1","deviceName":"Pixel"}`)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewReader(loginBody))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("login status=%d body=%s", w.Code, w.Body.String())
	}
	var loginResp struct {
		Token   string `json:"token"`
		Refresh string `json:"refresh"`
		User    struct {
			ID    int64  `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil || loginResp.Token == "" || loginResp.Refresh == "" {
		t.Fatalf("login parse: err=%v body=%s", err, w.Body.String())
	}

	// 3) Protected endpoint: wallets
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/wallets", nil)
	req.Header.Set("Authorization", "Bearer "+loginResp.Token)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("wallets list status=%d body=%s", w.Code, w.Body.String())
	}

	// 4) Refresh
	refBody := []byte(fmt.Sprintf(`{"refresh":"%s"}`, loginResp.Refresh))
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", bytes.NewReader(refBody))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("refresh status=%d body=%s", w.Code, w.Body.String())
	}
	var refResp struct {
		Token   string `json:"token"`
		Refresh string `json:"refresh"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &refResp); err != nil || refResp.Token == "" || refResp.Refresh == "" {
		t.Fatalf("refresh parse: err=%v body=%s", err, w.Body.String())
	}

	// 5) Health (public)
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("health status=%d body=%s", w.Code, w.Body.String())
	}
}
