package ports

import "time"

type Transaction struct {
	ID         int64      `db:"id"          json:"id"`
	UserID     int64      `db:"user_id"     json:"userId"`
	WalletID   int64      `db:"wallet_id"   json:"walletId"`
	CategoryID int64      `db:"category_id" json:"categoryId"`
	Type       string     `db:"type"        json:"type"`
	Amount     float64    `db:"amount"      json:"amount"`
	Currency   string     `db:"currency"    json:"currency"`
	Note       *string    `db:"note"        json:"note,omitempty"`
	OccurredAt time.Time  `db:"occurred_at" json:"occurredAt"`
	UpdatedAt  time.Time  `db:"updated_at"  json:"updatedAt"`
	DeletedAt  *time.Time `db:"deleted_at"  json:"deletedAt,omitempty"`
}

type TxSummary struct {
	Date  time.Time `db:"date"  json:"date"`
	Type  string    `db:"type"  json:"type"`
	Total float64   `db:"total" json:"total"`
}

type TxRepo interface {
	List(userID int64, page, size int, q string) ([]Transaction, int, error)
	ListRange(userID int64, from, to time.Time, q string, page, size int) ([]Transaction, int, error)
	GetSince(userID int64, since time.Time) ([]Transaction, error)
	UpsertBatch(userID int64, items []Transaction) error
	Create(userID int64, t *Transaction) error
	Update(userID int64, t *Transaction) error
	SoftDelete(userID int64, id int64) error
	Summary(userID int64, from, to time.Time) ([]TxSummary, error)
	GetOne(userID, id int64) (*Transaction, error)
}
