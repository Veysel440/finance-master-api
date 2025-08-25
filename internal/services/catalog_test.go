package services

import (
	"testing"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type fakeWalletRepo struct {
	lastCreate ports.Wallet
	lastUpdate ports.Wallet
	deleted    int64
}

func (r *fakeWalletRepo) List(int64) ([]ports.Wallet, error) { return nil, nil }
func (r *fakeWalletRepo) Create(_ int64, w *ports.Wallet) error {
	r.lastCreate = *w
	w.ID = 1
	return nil
}
func (r *fakeWalletRepo) Update(_ int64, w *ports.Wallet) error { r.lastUpdate = *w; return nil }
func (r *fakeWalletRepo) Delete(_ int64, id int64) error        { r.deleted = id; return nil }

type fakeCategoryRepo struct {
	lastCreate ports.Category
	lastUpdate ports.Category
	deleted    int64
}

func (r *fakeCategoryRepo) List(int64, string) ([]ports.Category, error) { return nil, nil }
func (r *fakeCategoryRepo) Create(_ int64, c *ports.Category) error {
	r.lastCreate = *c
	c.ID = 2
	return nil
}
func (r *fakeCategoryRepo) Update(_ int64, c *ports.Category) error { r.lastUpdate = *c; return nil }
func (r *fakeCategoryRepo) Delete(_ int64, id int64) error          { r.deleted = id; return nil }

func TestWallet_Create_DefaultCurrencyAndAudit(t *testing.T) {
	wr := &fakeWalletRepo{}
	ar := &fakeAuditRepo{}
	ws := &WalletService{Repo: wr, Audit: &AuditService{Repo: ar}}

	err := ws.Create(1, &ports.Wallet{Name: " Main "})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if wr.lastCreate.Currency != "TRY" {
		t.Fatalf("currency default failed")
	}
	if ar.last.action != "wallet.create" {
		t.Fatalf("audit not called")
	}
}

func TestWallet_Update_Validations(t *testing.T) {
	wr := &fakeWalletRepo{}
	ws := &WalletService{Repo: wr}
	if err := ws.Update(1, &ports.Wallet{ID: 1, Name: "   ", Currency: "TRY"}); err == nil {
		t.Fatalf("expected name_required")
	}
	if err := ws.Update(1, &ports.Wallet{ID: 1, Name: "OK", Currency: ""}); err == nil {
		t.Fatalf("expected currency_required")
	}
}

func TestCategory_Create_Validations(t *testing.T) {
	cr := &fakeCategoryRepo{}
	cs := &CategoryService{Repo: cr}
	if err := cs.Create(1, &ports.Category{Name: "", Type: "income"}); err == nil {
		t.Fatalf("expected name_required")
	}
	if err := cs.Create(1, &ports.Category{Name: "X", Type: "foo"}); err == nil {
		t.Fatalf("expected bad_type")
	}
}

func TestCategory_Update_Audit(t *testing.T) {
	cr := &fakeCategoryRepo{}
	ar := &fakeAuditRepo{}
	cs := &CategoryService{Repo: cr, Audit: &AuditService{Repo: ar}}
	err := cs.Update(1, &ports.Category{ID: 5, Name: "Market", Type: "expense"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ar.last.action != "category.update" {
		t.Fatalf("audit not called")
	}
}
