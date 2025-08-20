package ports

import "time"

type Transaction struct {
	ID         int64      `db:"id"`
	UserID     int64      `db:"user_id"`
	WalletID   int64      `db:"wallet_id"`
	CategoryID int64      `db:"category_id"`
	Type       string     `db:"type"`
	Amount     float64    `db:"amount"`
	Currency   string     `db:"currency"`
	Note       *string    `db:"note"`
	OccurredAt time.Time  `db:"occurred_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
	DeletedAt  *time.Time `db:"deleted_at"`
}

type TxSummary struct {
	Date  time.Time `db:"date" json:"date"`
	Type  string    `db:"type" json:"type"`
	Total float64   `db:"total" json:"total"`
}

type TxRepo interface {
	List(userID int64, page, size int, q string) ([]Transaction, int, error)
	GetSince(userID int64, since time.Time) ([]Transaction, error)
	UpsertBatch(userID int64, items []Transaction) error
	Create(userID int64, t *Transaction) error
	Update(userID int64, t *Transaction) error
	SoftDelete(userID int64, id int64) error
	Summary(userID int64, from, to time.Time) ([]TxSummary, error)
}
