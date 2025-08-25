package mysql

import (
	"context"
	"database/sql"

	"github.com/Veysel440/finance-master-api/internal/ports"
)

type IdempotencyRepo struct{ DB *sql.DB }

func NewIdempotencyRepo(db *sql.DB) *IdempotencyRepo { return &IdempotencyRepo{DB: db} }

func (r *IdempotencyRepo) Get(userID int64, key, resource string) (int64, bool, error) {
	var id int64
	err := r.DB.QueryRowContext(context.Background(),
		`SELECT resource_id FROM idempotency_keys WHERE user_id=? AND idem_key=? AND resource=?`,
		userID, key, resource,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return id, true, nil
}

func (r *IdempotencyRepo) Save(userID int64, key, resource string, resourceID int64) error {
	_, err := r.DB.ExecContext(context.Background(),
		`INSERT IGNORE INTO idempotency_keys (user_id, idem_key, resource, resource_id) VALUES (?,?,?,?)`,
		userID, key, resource, resourceID,
	)
	return err
}

var _ ports.IdempotencyRepo = (*IdempotencyRepo)(nil)
