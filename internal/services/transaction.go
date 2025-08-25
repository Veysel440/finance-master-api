package services

import (
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type TxService struct {
	Repo  ports.TxRepo
	Audit *AuditService
}

func (s *TxService) Create(uid int64, t *ports.Transaction) error {
	if err := s.Repo.Create(uid, t); err != nil {
		return err
	}
	s.alog(uid, "tx.create", t)
	return nil
}

func (s *TxService) Update(uid int64, t *ports.Transaction) error {
	if err := s.Repo.Update(uid, t); err != nil {
		return err
	}
	s.alog(uid, "tx.update", t)
	return nil
}

func (s *TxService) Delete(uid, id int64) error {
	if err := s.Repo.SoftDelete(uid, id); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(uid, "tx.delete", "transaction", &id, nil)
	}
	return nil
}

func (s *TxService) UpsertBatch(uid int64, items []ports.Transaction) error {
	if err := s.Repo.UpsertBatch(uid, items); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(uid, "tx.upsert_batch", "transaction", nil, map[string]any{"count": len(items)})
	}
	return nil
}

func (s *TxService) List(uid int64, page, size int, q string) ([]ports.Transaction, int, error) {
	return s.Repo.List(uid, page, size, q)
}

func (s *TxService) Summary(uid int64, from, to time.Time) ([]ports.TxSummary, error) {
	return s.Repo.Summary(uid, from, to)
}

func (s *TxService) GetOne(uid, id int64) (*ports.Transaction, error) {
	return s.Repo.GetOne(uid, id)
}

func (s *TxService) ListRange(uid int64, from, to time.Time, q string, page, size int) ([]ports.Transaction, int, error) {
	return s.Repo.ListRange(uid, from, to, q, page, size)
}
func (s *TxService) Since(uid int64, since time.Time) ([]ports.Transaction, error) {
	return s.Repo.GetSince(uid, since)
}

func (s *TxService) alog(uid int64, act string, t *ports.Transaction) {
	if s.Audit == nil || t == nil {
		return
	}
	s.Audit.Log(uid, act, "transaction", &t.ID, map[string]any{
		"amount":   t.Amount,
		"currency": t.Currency,
		"type":     t.Type,
	})
}
