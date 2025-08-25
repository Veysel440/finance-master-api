package services

import (
	"testing"
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type fakeTxRepo struct {
	created *ports.Transaction
	updated *ports.Transaction
	deleted int64
	batch   int
}

func (r *fakeTxRepo) Create(uid int64, t *ports.Transaction) error { r.created = t; return nil }
func (r *fakeTxRepo) Update(uid int64, t *ports.Transaction) error { r.updated = t; return nil }
func (r *fakeTxRepo) SoftDelete(uid, id int64) error               { r.deleted = id; return nil }
func (r *fakeTxRepo) UpsertBatch(uid int64, items []ports.Transaction) error {
	r.batch = len(items)
	return nil
}
func (r *fakeTxRepo) List(int64, int, int, string) ([]ports.Transaction, int, error) {
	return nil, 0, nil
}
func (r *fakeTxRepo) Summary(int64, time.Time, time.Time) ([]ports.TxSummary, error) { return nil, nil }
func (r *fakeTxRepo) GetOne(int64, int64) (*ports.Transaction, error)                { return nil, nil }
func (r *fakeTxRepo) ListRange(int64, time.Time, time.Time, string, int, int) ([]ports.Transaction, int, error) {
	return nil, 0, nil
}
func (r *fakeTxRepo) GetSince(int64, time.Time) ([]ports.Transaction, error) {
	return []ports.Transaction{}, nil
}

func TestTx_Create_Logs(t *testing.T) {
	txr := &fakeTxRepo{}
	ar := &fakeAuditRepo{}
	a := &AuditService{Repo: ar}
	svc := &TxService{Repo: txr, Audit: a}

	txx := &ports.Transaction{ID: 10, Amount: 12.3, Currency: "TRY", Type: "expense"}
	if err := svc.Create(9, txx); err != nil {
		t.Fatalf("err: %v", err)
	}
	if txr.created == nil || txr.created.ID != 10 {
		t.Fatalf("repo not called")
	}
	if ar.last.action != "tx.create" {
		t.Fatalf("audit not logged")
	}
}

func TestTx_Delete_Logs(t *testing.T) {
	txr := &fakeTxRepo{}
	ar := &fakeAuditRepo{}
	a := &AuditService{Repo: ar}
	svc := &TxService{Repo: txr, Audit: a}

	if err := svc.Delete(1, 77); err != nil {
		t.Fatalf("err: %v", err)
	}
	if txr.deleted != 77 {
		t.Fatalf("delete not called")
	}
	if ar.last.action != "tx.delete" {
		t.Fatalf("audit not logged")
	}
}

func TestTx_UpsertBatch_Count(t *testing.T) {
	txr := &fakeTxRepo{}
	ar := &fakeAuditRepo{}
	a := &AuditService{Repo: ar}
	svc := &TxService{Repo: txr, Audit: a}

	if err := svc.UpsertBatch(1, []ports.Transaction{{}, {}, {}}); err != nil {
		t.Fatalf("err: %v", err)
	}
	if txr.batch != 3 {
		t.Fatalf("batch size mismatch")
	}
}
