package services

import (
	"errors"
	"strings"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type WalletService struct{ Repo ports.WalletRepo }

func (s *WalletService) List(uid int64) ([]ports.Wallet, error) { return s.Repo.List(uid) }
func (s *WalletService) Create(uid int64, w *ports.Wallet) error {
	w.Name = strings.TrimSpace(w.Name)
	if w.Name == "" {
		return errors.New("name_required")
	}
	if w.Currency == "" {
		w.Currency = "TRY"
	}
	return s.Repo.Create(uid, w)
}
func (s *WalletService) Update(uid int64, w *ports.Wallet) error {
	w.Name = strings.TrimSpace(w.Name)
	if w.Name == "" {
		return errors.New("name_required")
	}
	if w.Currency == "" {
		return errors.New("currency_required")
	}
	return s.Repo.Update(uid, w)
}
func (s *WalletService) Delete(uid, id int64) error { return s.Repo.Delete(uid, id) }

type CategoryService struct{ Repo ports.CategoryRepo }

func (s *CategoryService) List(uid int64, typ string) ([]ports.Category, error) {
	return s.Repo.List(uid, typ)
}
func (s *CategoryService) Create(uid int64, c *ports.Category) error {
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return errors.New("name_required")
	}
	if c.Type != "income" && c.Type != "expense" {
		return errors.New("bad_type")
	}
	return s.Repo.Create(uid, c)
}
func (s *CategoryService) Update(uid int64, c *ports.Category) error {
	if c.ID == 0 {
		return errors.New("id_required")
	}
	c.Name = strings.TrimSpace(c.Name)
	if c.Name == "" {
		return errors.New("name_required")
	}
	if c.Type != "income" && c.Type != "expense" {
		return errors.New("bad_type")
	}
	return s.Repo.Update(uid, c)
}
func (s *CategoryService) Delete(uid, id int64) error { return s.Repo.Delete(uid, id) }
