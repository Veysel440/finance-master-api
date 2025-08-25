package services

import "github.com/Veysel440/finance-master-api/internal/ports"

type OnboardService struct {
	Wallet ports.WalletRepo
	Cat    ports.CategoryRepo
}

func (o *OnboardService) Seed(userID int64) error {
	if o == nil {
		return nil
	}
	if o.Wallet != nil {
		_ = o.Wallet.Create(userID, &ports.Wallet{Name: "Ana Cüzdan", Currency: "TRY"})
	}
	if o.Cat != nil {
		_ = o.Cat.Create(userID, &ports.Category{Name: "Maaş", Type: "income"})
		_ = o.Cat.Create(userID, &ports.Category{Name: "Serbest", Type: "income"})
		_ = o.Cat.Create(userID, &ports.Category{Name: "Yemek", Type: "expense"})
		_ = o.Cat.Create(userID, &ports.Category{Name: "Market", Type: "expense"})
	}
	return nil
}
