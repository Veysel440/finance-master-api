package services

import (
	"errors"
	"strings"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type WalletService struct {
	Repo  ports.WalletRepo
	Audit *AuditService
}

func (s *WalletService) List(uid int64) ([]ports.Wallet, error) { return s.Repo.List(uid) }

func (s *WalletService) Create(uid int64, w *ports.Wallet) error {
	w.Name = strings.TrimSpace(w.Name)
	if w.Name == "" {
		return errors.New("name_required")
	}
	if w.Currency == "" {
		w.Currency = "TRY"
	}
	if err := s.Repo.Create(uid, w); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(uid, "wallet.create", "wallet", &w.ID, map[string]any{"name": w.Name, "currency": w.Currency})
	}
	return nil
}
func (s *WalletService) Update(uid int64, w *ports.Wallet) error {
	w.Name = strings.TrimSpace(w.Name)
	if w.Name == "" {
		return errors.New("name_required")
	}
	if w.Currency == "" {
		return errors.New("currency_required")
	}
	if err := s.Repo.Update(uid, w); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(uid, "wallet.update", "wallet", &w.ID, map[string]any{"name": w.Name, "currency": w.Currency})
	}
	return nil
}
func (s *WalletService) Delete(uid, id int64) error {
	if err := s.Repo.Delete(uid, id); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(uid, "wallet.delete", "wallet", &id, nil)
	}
	return nil
}

type CategoryService struct {
	Repo  ports.CategoryRepo
	Audit *AuditService
}

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
	if err := s.Repo.Create(uid, c); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(uid, "category.create", "category", &c.ID, map[string]any{"name": c.Name, "type": c.Type})
	}
	return nil
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
	if err := s.Repo.Update(uid, c); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(uid, "category.update", "category", &c.ID, map[string]any{"name": c.Name, "type": c.Type})
	}
	return nil
}
func (s *CategoryService) Delete(uid, id int64) error {
	if err := s.Repo.Delete(uid, id); err != nil {
		return err
	}
	if s.Audit != nil {
		s.Audit.Log(uid, "category.delete", "category", &id, nil)
	}
	return nil
}
