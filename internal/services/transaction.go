package services

import (
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type TxService struct{ Repo ports.TxRepo }

func (s *TxService) List(uid int64, page, size int, q string) ([]ports.Transaction, int, error) {
	return s.Repo.List(uid, page, size, q)
}
func (s *TxService) Create(uid int64, t *ports.Transaction) error {
	t.UserID = uid
	t.UpdatedAt = time.Now()
	return s.Repo.Create(uid, t)
}
func (s *TxService) Update(uid int64, t *ports.Transaction) error {
	t.UserID = uid
	t.UpdatedAt = time.Now()
	return s.Repo.Update(uid, t)
}
func (s *TxService) Delete(uid, id int64) error { return s.Repo.SoftDelete(uid, id) }
func (s *TxService) Since(uid int64, since time.Time) ([]ports.Transaction, error) {
	return s.Repo.GetSince(uid, since)
}
func (s *TxService) UpsertBatch(uid int64, items []ports.Transaction) error {
	for i := range items {
		items[i].UserID = uid
		items[i].UpdatedAt = time.Now()
	}
	return s.Repo.UpsertBatch(uid, items)
}

func (s *TxService) Summary(uid int64, from, to time.Time) ([]ports.TxSummary, error) {
	return s.Repo.Summary(uid, from, to)
}
