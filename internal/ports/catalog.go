package ports

type Wallet struct {
	ID       int64  `db:"id" json:"id"`
	UserID   int64  `db:"user_id" json:"-"`
	Name     string `db:"name" json:"name"`
	Currency string `db:"currency" json:"currency"`
}
type Category struct {
	ID     int64  `db:"id" json:"id"`
	UserID int64  `db:"user_id" json:"-"`
	Name   string `db:"name" json:"name"`
	Type   string `db:"type" json:"type"`
}

type WalletRepo interface {
	List(userID int64) ([]Wallet, error)
	Create(userID int64, w *Wallet) error
	Update(userID int64, w *Wallet) error
	Delete(userID int64, id int64) error
}
type CategoryRepo interface {
	List(userID int64, typ string) ([]Category, error)
	Create(userID int64, c *Category) error
	Update(userID int64, c *Category) error
	Delete(userID int64, id int64) error
}
