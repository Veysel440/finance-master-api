package mysql

import (
	"strings"
	"time"

	"github.com/Veysel440/finance-master-api/internal/ports"
	"github.com/jmoiron/sqlx"
)

type TxRepo struct{ db *sqlx.DB }

func NewTxRepo(db *sqlx.DB) *TxRepo { return &TxRepo{db: db} }

func (r *TxRepo) List(userID int64, page, size int, q string) ([]ports.Transaction, int, error) {
	if size <= 0 {
		size = 20
	}
	if page <= 0 {
		page = 1
	}
	off := (page - 1) * size

	args := []any{userID}
	where := `WHERE t.user_id=? AND t.deleted_at IS NULL`

	if s := strings.TrimSpace(q); s != "" {
		if strings.HasPrefix(s, "ft:") {
			q = strings.TrimSpace(strings.TrimPrefix(s, "ft:"))
			where += ` AND MATCH(t.note) AGAINST (? IN NATURAL LANGUAGE MODE)`
			args = append(args, q)
		} else {
			where += ` AND (t.note LIKE ?)`
			args = append(args, "%"+s+"%")
		}
	}

	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM transactions t `+where, args...); err != nil {
		return nil, 0, err
	}

	args2 := append(args, size, off)
	rows := []ports.Transaction{}
	err := r.db.Select(&rows, `
		SELECT id, user_id, wallet_id, category_id, type, amount, currency, note,
		       occurred_at, updated_at, deleted_at
		FROM transactions t
		`+where+`
		ORDER BY occurred_at DESC, id DESC
		LIMIT ? OFFSET ?`, args2...)
	return rows, total, err
}

func (r *TxRepo) ListRange(userID int64, from, to time.Time, q string, page, size int) ([]ports.Transaction, int, error) {
	if size <= 0 {
		size = 500
	}
	if page <= 0 {
		page = 1
	}
	off := (page - 1) * size

	args := []any{userID, from, to}
	where := `WHERE t.user_id=? AND t.deleted_at IS NULL AND t.occurred_at >= ? AND t.occurred_at < ?`

	if s := strings.TrimSpace(q); s != "" {
		if strings.HasPrefix(s, "ft:") {
			q = strings.TrimSpace(strings.TrimPrefix(s, "ft:"))
			where += ` AND MATCH(t.note) AGAINST (? IN NATURAL LANGUAGE MODE)`
			args = append(args, q)
		} else {
			where += ` AND (t.note LIKE ?)`
			args = append(args, "%"+s+"%")
		}
	}

	var total int
	if err := r.db.Get(&total, `SELECT COUNT(*) FROM transactions t `+where, args...); err != nil {
		return nil, 0, err
	}

	args2 := append(args, size, off)
	rows := []ports.Transaction{}
	err := r.db.Select(&rows, `
		SELECT id, user_id, wallet_id, category_id, type, amount, currency, note,
		       occurred_at, updated_at, deleted_at
		FROM transactions t
		`+where+`
		ORDER BY occurred_at ASC, id ASC
		LIMIT ? OFFSET ?`, args2...)
	return rows, total, err
}

func (r *TxRepo) GetSince(userID int64, since time.Time) ([]ports.Transaction, error) {
	rows := []ports.Transaction{}
	err := r.db.Select(&rows, `
		SELECT id, user_id, wallet_id, category_id, type, amount, currency, note,
		       occurred_at, updated_at, deleted_at
		FROM transactions
		WHERE user_id=? AND (updated_at > ? OR (deleted_at IS NOT NULL AND deleted_at > ?))
		ORDER BY updated_at ASC`, userID, since, since)
	return rows, err
}

func (r *TxRepo) Create(userID int64, t *ports.Transaction) error {
	res, err := r.db.Exec(`
		INSERT INTO transactions
		(user_id, wallet_id, category_id, type, amount, currency, note, occurred_at, updated_at)
		VALUES (?,?,?,?,?,?,?,?,NOW())`,
		userID, t.WalletID, t.CategoryID, t.Type, t.Amount, t.Currency, t.Note, t.OccurredAt)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	t.ID = id
	return nil
}

func (r *TxRepo) Update(userID int64, t *ports.Transaction) error {
	_, err := r.db.Exec(`
		UPDATE transactions
		SET wallet_id=?, category_id=?, type=?, amount=?, currency=?, note=?, occurred_at=?, updated_at=NOW()
		WHERE id=? AND user_id=?`,
		t.WalletID, t.CategoryID, t.Type, t.Amount, t.Currency, t.Note, t.OccurredAt, t.ID, userID)
	return err
}

func (r *TxRepo) SoftDelete(userID int64, id int64) error {
	_, err := r.db.Exec(`
		UPDATE transactions
		SET deleted_at=NOW(), updated_at=NOW()
		WHERE id=? AND user_id=?`, id, userID)
	return err
}

func (r *TxRepo) UpsertBatch(userID int64, items []ports.Transaction) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else {
			_ = tx.Commit()
		}
	}()

	const ins = `INSERT INTO transactions
		(id, user_id, wallet_id, category_id, type, amount, currency, note, occurred_at, updated_at, deleted_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)
		ON DUPLICATE KEY UPDATE
		  wallet_id=VALUES(wallet_id),
		  category_id=VALUES(category_id),
		  type=VALUES(type),
		  amount=VALUES(amount),
		  currency=VALUES(currency),
		  note=VALUES(note),
		  occurred_at=VALUES(occurred_at),
		  updated_at=VALUES(updated_at),
		  deleted_at=VALUES(deleted_at)`

	for i := range items {
		it := items[i]
		if it.UpdatedAt.IsZero() {
			it.UpdatedAt = time.Now()
		}
		_, err = tx.Exec(ins,
			nullID(it.ID), userID, it.WalletID, it.CategoryID, it.Type, it.Amount, it.Currency, it.Note,
			it.OccurredAt, it.UpdatedAt, it.DeletedAt)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TxRepo) Summary(userID int64, from, to time.Time) ([]ports.TxSummary, error) {
	rows := []ports.TxSummary{}
	err := r.db.Select(&rows, `
		SELECT DATE(occurred_at) AS date, type, SUM(amount) AS total
		FROM transactions
		WHERE user_id=? AND deleted_at IS NULL AND occurred_at >= ? AND occurred_at < ?
		GROUP BY DATE(occurred_at), type
		ORDER BY DATE(occurred_at) ASC, type ASC`, userID, from, to)
	return rows, err
}

func (r *TxRepo) GetOne(userID, id int64) (*ports.Transaction, error) {
	var t ports.Transaction
	err := r.db.Get(&t, `
		SELECT id, user_id, wallet_id, category_id, type, amount, currency, note,
		       occurred_at, updated_at, deleted_at
		FROM transactions
		WHERE id=? AND user_id=? AND deleted_at IS NULL
		LIMIT 1`, id, userID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func nullID(id int64) any {
	if id > 0 {
		return id
	}
	return nil
}
