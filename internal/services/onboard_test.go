package services

import (
	"testing"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type fakeWalRepo struct{ created int }

func (f *fakeWalRepo) List(int64) ([]ports.Wallet, error) { return nil, nil }
func (f *fakeWalRepo) Create(int64, *ports.Wallet) error  { f.created++; return nil }
func (f *fakeWalRepo) Update(int64, *ports.Wallet) error  { return nil }
func (f *fakeWalRepo) Delete(int64, int64) error          { return nil }

type fakeCatRepo struct{ created int }

func (f *fakeCatRepo) List(int64, string) ([]ports.Category, error) { return nil, nil }
func (f *fakeCatRepo) Create(int64, *ports.Category) error          { f.created++; return nil }
func (f *fakeCatRepo) Update(int64, *ports.Category) error          { return nil }
func (f *fakeCatRepo) Delete(int64, int64) error                    { return nil }

func TestOnboard_Seed(t *testing.T) {
	wr, cr := &fakeWalRepo{}, &fakeCatRepo{}
	o := &OnboardService{Wallet: wr, Cat: cr}
	if err := o.Seed(1); err != nil {
		t.Fatalf("seed err: %v", err)
	}
	if wr.created < 1 || cr.created < 4 {
		t.Fatalf("defaults not created")
	}
}
